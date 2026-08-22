package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// matches / match_participants の CHECK 制約と同じ値。
const (
	gameStatusInProgress = "in_progress"
	gameStatusFinished   = "finished"
	gameStatusAborted    = "aborted"
	gameResultTypeAbort  = "abort"
)

// match_actions の制約名。冪等キー・版数の衝突をこの名前で見分ける。
const (
	uxMatchActionsActionID = "ux_match_actions_action_id"
	pkMatchActions         = "match_actions_pkey"
)

// GameRepo は進行中の対局の永続化。
type GameRepo struct {
	db *gorm.DB
}

func NewGameRepo(db *gorm.DB) *GameRepo {
	return &GameRepo{db: db}
}

// CreateMatch は対戦 1 件と参加者 2 行（未決着）を 1 トランザクションで作る。
// rating_before はこの時点の users.rating を写す。
func (r *GameRepo) CreateMatch(ctx context.Context, mode string, userIDs [2]uuid.UUID) (*usecase.StoredGame, error) {
	var stored *usecase.StoredGame
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var users []User
		if err := tx.Where("id IN ?", []uuid.UUID{userIDs[0], userIDs[1]}).Find(&users).Error; err != nil {
			return fmt.Errorf("find players: %w", err)
		}
		if len(users) != 2 {
			return domain.ErrUserNotFound
		}
		byID := map[uuid.UUID]*User{users[0].ID: &users[0], users[1].ID: &users[1]}

		match := Match{
			Mode:      mode,
			Status:    gameStatusInProgress,
			StartedAt: time.Now(),
		}
		if err := tx.Create(&match).Error; err != nil {
			return fmt.Errorf("create match: %w", err)
		}

		game := usecase.StoredGame{
			MatchID:   match.ID,
			Mode:      mode,
			StartedAt: match.StartedAt,
		}
		for seat, userID := range userIDs {
			u := byID[userID]
			participant := MatchParticipant{
				MatchID:      match.ID,
				UserID:       u.ID,
				Seat:         int16(seat), //nolint:gosec // seat は 0|1
				RatingBefore: u.Rating,
			}
			if err := tx.Create(&participant).Error; err != nil {
				return fmt.Errorf("create participant: %w", err)
			}
			game.Players[seat] = usecase.GamePlayer{
				UserID:      u.ID,
				DisplayName: u.DisplayName,
				Handle:      u.Handle,
				Rating:      u.Rating,
			}
		}
		stored = &game
		return nil
	})
	if err != nil {
		return nil, err
	}
	return stored, nil
}

// AppendAction は 1 手を追記する。ユニーク制約への衝突は domain のエラーに翻訳する。
func (r *GameRepo) AppendAction(ctx context.Context, record usecase.MatchActionRecord) error {
	action := MatchAction{
		MatchID:    record.MatchID,
		ActionSeq:  record.Seq,
		ActionID:   record.ActionID,
		ActorSeat:  int16(record.Seat), //nolint:gosec // seat は 0|1
		ActionType: record.Type,
		Payload:    string(record.Payload),
	}
	if err := r.db.WithContext(ctx).Create(&action).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			switch pgErr.ConstraintName {
			case uxMatchActionsActionID:
				return domain.ErrDuplicateGameAction // 同じ操作の再送
			case pkMatchActions:
				return domain.ErrStaleGameVersion // 同じ版数へ別の手が先に入った
			}
		}
		return fmt.Errorf("append action: %w", err)
	}
	return nil
}

// FinishMatch は対局を決着させ、勝敗・レーティング・XP を反映する。
// participants が空なら中断（aborted）として勝敗は付けずに閉じる。
// users の rating / XP / level の更新は RecordMatch と同じ updateUserProgress を使う。
func (r *GameRepo) FinishMatch(ctx context.Context, matchID uuid.UUID, resultType string, totalMoves int, participants []domain.MatchParticipant) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		status := gameStatusFinished
		if len(participants) == 0 {
			status = gameStatusAborted
			resultType = gameResultTypeAbort
		}
		res := tx.Model(&Match{}).
			Where("id = ? AND status = ?", matchID, gameStatusInProgress).
			Updates(map[string]any{
				"status":      status,
				"result_type": resultType,
				"total_moves": totalMoves,
				"finished_at": time.Now(),
			})
		if res.Error != nil {
			return fmt.Errorf("finish match: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return domain.ErrMatchNotFound // 既に決着済みか、存在しない
		}

		for _, p := range participants {
			err := tx.Model(&MatchParticipant{}).
				Where("match_id = ? AND seat = ?", matchID, p.Seat).
				Updates(map[string]any{
					"outcome":      p.Outcome,
					"rating_after": p.RatingAfter,
					"xp_gained":    p.XPGained,
				}).Error
			if err != nil {
				return fmt.Errorf("settle participant: %w", err)
			}
			if err := updateUserProgress(tx, p); err != nil {
				return err
			}
		}
		return nil
	})
}

// FindActiveMatch は user が参加している進行中の対局を、手のログごと読み出す。
func (r *GameRepo) FindActiveMatch(ctx context.Context, userID uuid.UUID) (*usecase.StoredGame, error) {
	var match Match
	err := r.db.WithContext(ctx).
		Joins("JOIN match_participants mp ON mp.match_id = matches.id").
		Where("mp.user_id = ? AND matches.status = ?", userID, gameStatusInProgress).
		Order("matches.started_at").
		First(&match).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMatchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find active match: %w", err)
	}
	return r.loadStored(ctx, &match)
}

// FindActiveMatchByID は進行中の対局を ID で引く。観戦の入り口。
func (r *GameRepo) FindActiveMatchByID(ctx context.Context, matchID uuid.UUID) (*usecase.StoredGame, error) {
	var match Match
	err := r.db.WithContext(ctx).
		First(&match, "id = ? AND status = ?", matchID, gameStatusInProgress).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMatchNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find match: %w", err)
	}
	return r.loadStored(ctx, &match)
}

// ListLiveMatches は進行中の対局を開始の新しい順に返す。
// 手数は match_actions の件数から数える（matches.total_moves は決着時にしか入らない）。
func (r *GameRepo) ListLiveMatches(ctx context.Context, limit, offset int) ([]usecase.LiveMatch, int, error) {
	base := r.db.WithContext(ctx).Model(&Match{}).Where("status = ?", gameStatusInProgress)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count live matches: %w", err)
	}

	var matches []Match
	err := base.Order("started_at DESC").Limit(limit).Offset(offset).Find(&matches).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list live matches: %w", err)
	}

	live := make([]usecase.LiveMatch, 0, len(matches))
	for i := range matches {
		stored, err := r.loadStored(ctx, &matches[i])
		if err != nil {
			return nil, 0, err
		}
		live = append(live, usecase.LiveMatch{
			MatchID:   stored.MatchID,
			Mode:      stored.Mode,
			StartedAt: stored.StartedAt,
			Players:   stored.Players,
			MoveCount: len(stored.Actions),
		})
	}
	return live, int(total), nil
}

func (r *GameRepo) loadStored(ctx context.Context, match *Match) (*usecase.StoredGame, error) {
	var participants []MatchParticipant
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("match_id = ?", match.ID).
		Order("seat").
		Find(&participants).Error
	if err != nil {
		return nil, fmt.Errorf("load participants: %w", err)
	}
	if len(participants) != 2 {
		return nil, fmt.Errorf("match %s has %d participants", match.ID, len(participants))
	}

	var actions []MatchAction
	err = r.db.WithContext(ctx).
		Where("match_id = ?", match.ID).
		Order("action_seq").
		Find(&actions).Error
	if err != nil {
		return nil, fmt.Errorf("load actions: %w", err)
	}

	stored := usecase.StoredGame{
		MatchID:   match.ID,
		Mode:      match.Mode,
		StartedAt: match.StartedAt,
	}
	for i, p := range participants {
		player := usecase.GamePlayer{
			UserID: p.UserID,
			Rating: p.RatingBefore,
		}
		if p.User != nil {
			player.DisplayName = p.User.DisplayName
			player.Handle = p.User.Handle
			player.Rating = p.User.Rating
		}
		stored.Players[i] = player
	}
	for _, a := range actions {
		stored.Actions = append(stored.Actions, usecase.StoredAction{
			Seq:      a.ActionSeq,
			ActionID: a.ActionID,
			Seat:     int(a.ActorSeat),
			Type:     a.ActionType,
			Payload:  []byte(a.Payload),
		})
	}
	return &stored, nil
}
