package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// matches.status のうち、統計が対象にするもの。進行中と中断は集計に入れない。
const matchStatusFinished = "finished"

// matchRecordRow は履歴の 1 行。自分の行と相手の行を JOIN した形で受ける。
// GORM のモデルではなくクエリ専用の型なので、テーブルとは対応しない。
type matchRecordRow struct {
	MatchID      uuid.UUID
	Mode         string
	ResultType   string
	TotalMoves   int
	StartedAt    time.Time
	FinishedAt   time.Time
	Outcome      string
	RatingBefore int
	RatingAfter  int
	XPGained     int

	OpponentID          uuid.UUID
	OpponentDisplayName string
	OpponentHandle      string
	OpponentLevel       int
}

func (row matchRecordRow) toDomain() domain.MatchRecord {
	return domain.MatchRecord{
		Match: domain.Match{
			ID:         row.MatchID,
			Mode:       row.Mode,
			ResultType: row.ResultType,
			TotalMoves: row.TotalMoves,
			StartedAt:  row.StartedAt,
			FinishedAt: row.FinishedAt,
		},
		Opponent: domain.User{
			ID:          row.OpponentID,
			DisplayName: row.OpponentDisplayName,
			Handle:      row.OpponentHandle,
			Level:       row.OpponentLevel,
		},
		Outcome:      row.Outcome,
		ResultType:   row.ResultType,
		RatingBefore: row.RatingBefore,
		RatingAfter:  row.RatingAfter,
		XPGained:     row.XPGained,
	}
}

// MatchRepo は対戦の永続化。
type MatchRepo struct {
	db *gorm.DB
}

func NewMatchRepo(db *gorm.DB) *MatchRepo {
	return &MatchRepo{db: db}
}

func (r *MatchRepo) UsersByID(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]domain.User, error) {
	var rows []User

	err := r.db.WithContext(ctx).
		Where("id IN ? AND anonymized_at IS NULL", userIDs).
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("select users: %w", err)
	}

	users := make(map[uuid.UUID]domain.User, len(rows))
	for _, row := range rows {
		users[row.ID] = *toDomainUser(&row)
	}

	return users, nil
}

// RecordMatch は対戦・参加者・各人の rating/XP/level をまとめて書く。
//
// 途中で失敗すると「対戦は残ったがレーティングは動いていない」といった
// 中途半端な状態になるので、1 トランザクションにまとめる。
func (r *MatchRepo) RecordMatch(ctx context.Context, match domain.Match) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := Match{
			ID:         match.ID,
			Mode:       match.Mode,
			Status:     matchStatusFinished,
			ResultType: &match.ResultType,
			TotalMoves: match.TotalMoves,
			StartedAt:  match.StartedAt,
			FinishedAt: &match.FinishedAt,
		}
		if err := tx.Create(&row).Error; err != nil {
			return fmt.Errorf("insert match: %w", err)
		}

		for _, p := range match.Participants {
			participant := MatchParticipant{
				MatchID: match.ID,
				UserID:  p.UserID,
				Seat:    int16(p.Seat),
				// 列は進行中の対局のために NULL 許容だが、ここは決着の記録なので必ず埋める。
				Outcome:      &p.Outcome,
				RatingBefore: p.RatingBefore,
				RatingAfter:  &p.RatingAfter,
				XPGained:     p.XPGained,
			}
			if err := tx.Create(&participant).Error; err != nil {
				return fmt.Errorf("insert match participant: %w", err)
			}

			if err := updateUserProgress(tx, p); err != nil {
				return err
			}
		}

		return nil
	})
}

// updateUserProgress は 1 人ぶんの rating / XP / level を進める。
//
// level は experience_points から毎回導出して上書きする。両方を独立に足していくと、
// 片方の更新が漏れたときに静かにズレるため。
// gorm.Expr を使うのは、読んでから書くまでの間に別の対戦が入っても取りこぼさないようにするため。
func updateUserProgress(tx *gorm.DB, p domain.MatchParticipant) error {
	err := tx.Model(&User{}).
		Where("id = ?", p.UserID).
		Updates(map[string]any{
			"rating":            p.RatingAfter,
			"experience_points": gorm.Expr("experience_points + ?", p.XPGained),
		}).Error
	if err != nil {
		return fmt.Errorf("update user progress: %w", err)
	}

	var user User
	if err := tx.Select("id", "experience_points").First(&user, "id = ?", p.UserID).Error; err != nil {
		return fmt.Errorf("select user xp: %w", err)
	}

	err = tx.Model(&User{}).
		Where("id = ?", p.UserID).
		Update("level", domain.LevelForXP(user.ExperiencePoints)).Error
	if err != nil {
		return fmt.Errorf("update user level: %w", err)
	}

	return nil
}

func (r *MatchRepo) ListMatches(
	ctx context.Context, userID uuid.UUID, f usecase.MatchFilter,
) ([]domain.MatchRecord, error) {
	var rows []matchRecordRow

	err := r.matchQuery(ctx, userID, f).
		Select(`m.id AS match_id, m.mode, m.result_type, m.total_moves,
			m.started_at, m.finished_at,
			mp.outcome, mp.rating_before, mp.rating_after, mp.xp_gained,
			u.id AS opponent_id, u.display_name AS opponent_display_name,
			u.handle AS opponent_handle, u.level AS opponent_level`).
		Order("m.finished_at DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("select matches: %w", err)
	}

	records := make([]domain.MatchRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, row.toDomain())
	}

	return records, nil
}

func (r *MatchRepo) CountMatches(ctx context.Context, userID uuid.UUID, f usecase.MatchFilter) (int, error) {
	var total int64

	if err := r.matchQuery(ctx, userID, f).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count matches: %w", err)
	}

	return int(total), nil
}

// matchQuery は履歴と件数で共通の絞り込み。
//
// 自分の行 (mp) と相手の行 (opp) を同じ match_id で突き合わせる。
// match_participants は 1 対戦 2 行なので、mp.user_id <> opp.user_id で相手が一意に決まる。
func (r *MatchRepo) matchQuery(ctx context.Context, userID uuid.UUID, f usecase.MatchFilter) *gorm.DB {
	q := r.db.WithContext(ctx).
		Table("match_participants AS mp").
		Joins("JOIN matches AS m ON m.id = mp.match_id").
		Joins("JOIN match_participants AS opp ON opp.match_id = mp.match_id AND opp.user_id <> mp.user_id").
		Joins("JOIN users AS u ON u.id = opp.user_id").
		Where("mp.user_id = ? AND m.status = ?", userID, matchStatusFinished)

	if f.From != nil {
		q = q.Where("m.finished_at >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("m.finished_at <= ?", *f.To)
	}
	if f.Mode != "" {
		q = q.Where("m.mode = ?", f.Mode)
	}
	if f.Outcome != "" {
		q = q.Where("mp.outcome = ?", f.Outcome)
	}

	return q
}
