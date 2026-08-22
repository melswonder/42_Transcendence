package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// StatsRepo は統計の集計。集計結果のキャッシュ表は持たず、毎回クエリで求める。
// 二重管理でズレるより、インデックスを効かせて都度数えるほうが確実なため。
type StatsRepo struct {
	db *gorm.DB
}

func NewStatsRepo(db *gorm.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

var _ usecase.StatsRepository = (*StatsRepo)(nil)

// statsWhere は集計クエリ共通の絞り込みを、条件文と引数に組み立てる。
//
// 生 SQL を使うのは、FILTER 句・date_trunc・窓関数が GORM のビルダーでは
// かえって読みにくくなるため。プレースホルダは必ず使い、値は連結しない。
func statsWhere(userID uuid.UUID, f usecase.MatchFilter) (string, []any) {
	conds := []string{"mp.user_id = ?", "m.status = ?"}
	args := []any{userID, matchStatusFinished}

	if f.From != nil {
		conds = append(conds, "m.finished_at >= ?")
		args = append(args, *f.From)
	}
	if f.To != nil {
		conds = append(conds, "m.finished_at <= ?")
		args = append(args, *f.To)
	}
	if f.Mode != "" {
		conds = append(conds, "m.mode = ?")
		args = append(args, f.Mode)
	}
	if f.Outcome != "" {
		conds = append(conds, "mp.outcome = ?")
		args = append(args, f.Outcome)
	}
	if f.Opponent != nil {
		// 集計側のクエリは opp を JOIN していないので EXISTS で相手を絞る。
		conds = append(conds, `EXISTS (
			SELECT 1 FROM match_participants opp
			WHERE opp.match_id = m.id AND opp.user_id = ?
		)`)
		args = append(args, *f.Opponent)
	}

	return strings.Join(conds, " AND "), args
}

func (r *StatsRepo) Summary(
	ctx context.Context, userID uuid.UUID, f usecase.MatchFilter,
) (domain.StatsSummary, error) {
	where, args := statsWhere(userID, f)

	var counts struct {
		Wins   int
		Losses int
		Draws  int
	}

	query := fmt.Sprintf(`
		SELECT
			count(*) FILTER (WHERE mp.outcome = 'win')  AS wins,
			count(*) FILTER (WHERE mp.outcome = 'loss') AS losses,
			count(*) FILTER (WHERE mp.outcome = 'draw') AS draws
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		WHERE %s`, where)

	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&counts).Error; err != nil {
		return domain.StatsSummary{}, fmt.Errorf("select stats summary: %w", err)
	}

	// レーティング・レベル・順位は期間で絞らない。
	// 「今の自分」を出す欄なので、日付範囲を変えても動かないのが正しい。
	var me struct {
		Rating       int
		Level        int
		XP           int
		Ranking      int
		TotalPlayers int
	}

	err := r.db.WithContext(ctx).Raw(`
		SELECT u.rating, u.level, u.experience_points AS xp,
			(SELECT count(*) + 1 FROM users o
			 WHERE o.rating > u.rating AND o.anonymized_at IS NULL AND o.status = 'active') AS ranking,
			(SELECT count(*) FROM users o
			 WHERE o.anonymized_at IS NULL AND o.status = 'active') AS total_players
		FROM users u
		WHERE u.id = ?`, userID).Scan(&me).Error
	if err != nil {
		return domain.StatsSummary{}, fmt.Errorf("select user standing: %w", err)
	}

	return domain.StatsSummary{
		Wins:         counts.Wins,
		Losses:       counts.Losses,
		Draws:        counts.Draws,
		Rating:       me.Rating,
		Ranking:      me.Ranking,
		TotalPlayers: me.TotalPlayers,
		Level:        me.Level,
		XP:           me.XP,
	}, nil
}

// Outcomes は勝敗だけを古い順に返す。連勝の判定は domain 側で行う。
func (r *StatsRepo) Outcomes(
	ctx context.Context, userID uuid.UUID, f usecase.MatchFilter,
) ([]string, error) {
	where, args := statsWhere(userID, f)

	var outcomes []string

	query := fmt.Sprintf(`
		SELECT mp.outcome
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		WHERE %s
		ORDER BY m.finished_at ASC`, where)

	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&outcomes).Error; err != nil {
		return nil, fmt.Errorf("select outcomes: %w", err)
	}

	return outcomes, nil
}

func (r *StatsRepo) Timeseries(
	ctx context.Context, userID uuid.UUID, f usecase.MatchFilter, interval string,
) ([]domain.TimeseriesPoint, error) {
	where, args := statsWhere(userID, f)

	// interval は domain.NormalizeInterval で既知の値に丸めてから来るので、
	// ここで文字列として埋め込んでも注入にはならない。念のため再度絞る。
	trunc := "day"
	if interval == domain.IntervalWeek {
		trunc = "week"
	}

	var rows []struct {
		Date    time.Time
		Wins    int
		Losses  int
		Draws   int
		Matches int
		Rating  int
	}

	// rating は「そのコマの最後の対戦を終えた時点」を採る。
	// 平均にすると推移が鈍るし、最初の値だと当日の伸びが見えないため。
	query := fmt.Sprintf(`
		SELECT date_trunc('%s', m.finished_at) AS date,
			count(*) FILTER (WHERE mp.outcome = 'win')  AS wins,
			count(*) FILTER (WHERE mp.outcome = 'loss') AS losses,
			count(*) FILTER (WHERE mp.outcome = 'draw') AS draws,
			count(*) AS matches,
			(array_agg(mp.rating_after ORDER BY m.finished_at DESC))[1] AS rating
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		WHERE %s
		GROUP BY date
		ORDER BY date ASC`, trunc, where)

	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("select timeseries: %w", err)
	}

	points := make([]domain.TimeseriesPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, domain.TimeseriesPoint{
			Date:    row.Date,
			Wins:    row.Wins,
			Losses:  row.Losses,
			Draws:   row.Draws,
			Matches: row.Matches,
			Rating:  row.Rating,
		})
	}

	return points, nil
}

func (r *StatsRepo) Breakdown(
	ctx context.Context, userID uuid.UUID, f usecase.MatchFilter,
) (domain.Breakdown, error) {
	byResultType, err := r.sliceBy(ctx, userID, f, "m.result_type")
	if err != nil {
		return domain.Breakdown{}, err
	}

	byMode, err := r.sliceBy(ctx, userID, f, "m.mode")
	if err != nil {
		return domain.Breakdown{}, err
	}

	byOutcome, err := r.sliceBy(ctx, userID, f, "mp.outcome")
	if err != nil {
		return domain.Breakdown{}, err
	}

	return domain.Breakdown{
		ByResultType: byResultType,
		ByMode:       byMode,
		ByOutcome:    byOutcome,
	}, nil
}

// sliceBy は列 1 つで GROUP BY した件数を返す。
// column は呼び出し側が固定文字列で渡す（クエリから来た値は入らない）。
func (r *StatsRepo) sliceBy(
	ctx context.Context, userID uuid.UUID, f usecase.MatchFilter, column string,
) ([]domain.BreakdownSlice, error) {
	where, args := statsWhere(userID, f)

	var rows []domain.BreakdownSlice

	query := fmt.Sprintf(`
		SELECT %s AS key, count(*) AS count
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id
		WHERE %s
		GROUP BY key
		ORDER BY count DESC, key ASC`, column, where)

	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("select breakdown by %s: %w", column, err)
	}

	return rows, nil
}

// leaderboardSelect はランキング 1 行ぶんの列。順位・自分の行の両方で使う。
//
// 順位は RANK() で毎回求める。順位を列として持つと、
// 誰か 1 人のレーティングが動くたびに他の行も書き換えることになるため。
const leaderboardSelect = `
	SELECT u.id, u.display_name, u.handle, u.level, u.rating,
		RANK() OVER (ORDER BY u.rating DESC) AS rank,
		COALESCE(s.wins, 0)   AS wins,
		COALESCE(s.losses, 0) AS losses
	FROM users u
	LEFT JOIN (
		SELECT mp.user_id,
			count(*) FILTER (WHERE mp.outcome = 'win')  AS wins,
			count(*) FILTER (WHERE mp.outcome = 'loss') AS losses
		FROM match_participants mp
		JOIN matches m ON m.id = mp.match_id AND m.status = 'finished'
		GROUP BY mp.user_id
	) s ON s.user_id = u.id
	WHERE u.anonymized_at IS NULL AND u.status = 'active'`

type leaderboardRow struct {
	ID          uuid.UUID
	DisplayName string
	Handle      string
	Level       int
	Rating      int
	Rank        int
	Wins        int
	Losses      int
}

func (row leaderboardRow) toDomain() domain.LeaderboardEntry {
	return domain.LeaderboardEntry{
		Rank: row.Rank,
		User: domain.User{
			ID:          row.ID,
			DisplayName: row.DisplayName,
			Handle:      row.Handle,
			Level:       row.Level,
			Rating:      row.Rating,
		},
		Rating: row.Rating,
		Wins:   row.Wins,
		Losses: row.Losses,
	}
}

func (r *StatsRepo) Leaderboard(
	ctx context.Context, limit, offset int,
) ([]domain.LeaderboardEntry, int, error) {
	var total int64

	err := r.db.WithContext(ctx).
		Model(&User{}).
		Where("anonymized_at IS NULL AND status = ?", userStatusActive).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("count players: %w", err)
	}

	var rows []leaderboardRow

	// 同率のときの並びを handle で固定する。付けないと呼ぶたびに順番が変わりうる。
	query := leaderboardSelect + `
		ORDER BY u.rating DESC, u.handle ASC
		LIMIT ? OFFSET ?`

	if err := r.db.WithContext(ctx).Raw(query, limit, offset).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("select leaderboard: %w", err)
	}

	entries := make([]domain.LeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, row.toDomain())
	}

	return entries, int(total), nil
}

// LeaderboardEntryOf は 1 人ぶんの順位を引く。
//
// RANK() は WHERE より後に評価されるので、先に全員ぶんを数えてから絞る必要がある。
// そのため CTE で包んでいる。
func (r *StatsRepo) LeaderboardEntryOf(
	ctx context.Context, userID uuid.UUID,
) (*domain.LeaderboardEntry, error) {
	var row leaderboardRow

	query := `WITH ranked AS (` + leaderboardSelect + `)
		SELECT * FROM ranked WHERE id = ?`

	err := r.db.WithContext(ctx).Raw(query, userID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select my rank: %w", err)
	}

	entry := row.toDomain()

	return &entry, nil
}

