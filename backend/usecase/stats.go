package usecase

import (
	"context"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// StatsRepository は統計の集計。すべて読み取り専用。
// Summary は勝敗・レーティング・順位・レベルをまとめて引く。
// Outcomes は集計対象の勝敗を古い順に返す（連勝の計算に使う）。
// Timeseries は日別・週別に丸めた集計を返す。
// Breakdown は結果種別・モード・勝敗ごとの件数を返す。
// Leaderboard はレーティング上位を返す。
// LeaderboardEntryOf は 1 人ぶんの順位を引く。表示範囲の外にいても順位を出すため。
type StatsRepository interface {
	Summary(ctx context.Context, userID uuid.UUID, f MatchFilter) (domain.StatsSummary, error)
	Outcomes(ctx context.Context, userID uuid.UUID, f MatchFilter) ([]string, error)
	Timeseries(ctx context.Context, userID uuid.UUID, f MatchFilter, interval string) ([]domain.TimeseriesPoint, error)
	Breakdown(ctx context.Context, userID uuid.UUID, f MatchFilter) (domain.Breakdown, error)
	Leaderboard(ctx context.Context, limit, offset int) ([]domain.LeaderboardEntry, int, error)
	LeaderboardEntryOf(ctx context.Context, userID uuid.UUID) (*domain.LeaderboardEntry, error)
	// Opponents は対戦したことのある相手の一覧。フィルタの選択肢用。
	Opponents(ctx context.Context, userID uuid.UUID) ([]domain.User, error)
}

// StatsUsecase は統計の取得を進める。書き込みは一切しない。
type StatsUsecase struct {
	repo StatsRepository
}

func NewStatsUsecase(repo StatsRepository) *StatsUsecase {
	return &StatsUsecase{repo: repo}
}

// Summary は数値タイルぶんの集計を返す。
//
// 連勝は SQL で数えず、勝敗の列だけ取ってきて domain 側で数える。
// 窓関数で書けなくはないが、クエリが読みにくくなるうえテストしづらいため。
func (u *StatsUsecase) Summary(
	ctx context.Context, userID uuid.UUID, f MatchFilter,
) (domain.StatsSummary, error) {
	summary, err := u.repo.Summary(ctx, userID, f)
	if err != nil {
		return domain.StatsSummary{}, err
	}

	outcomes, err := u.repo.Outcomes(ctx, userID, f)
	if err != nil {
		return domain.StatsSummary{}, err
	}

	summary.CurrentStreak, summary.BestStreak = domain.Streaks(outcomes)

	return summary, nil
}

func (u *StatsUsecase) Timeseries(
	ctx context.Context, userID uuid.UUID, f MatchFilter, interval string,
) ([]domain.TimeseriesPoint, error) {
	return u.repo.Timeseries(ctx, userID, f, domain.NormalizeInterval(interval))
}

func (u *StatsUsecase) Breakdown(
	ctx context.Context, userID uuid.UUID, f MatchFilter,
) (domain.Breakdown, error) {
	return u.repo.Breakdown(ctx, userID, f)
}

// Leaderboard は上位一覧と、その中での自分の位置を返す。
// 自分が表示範囲の外にいても順位を出せるよう、me は別に引く。
func (u *StatsUsecase) Leaderboard(
	ctx context.Context, userID uuid.UUID, limit, offset int,
) ([]domain.LeaderboardEntry, *domain.LeaderboardEntry, int, error) {
	entries, total, err := u.repo.Leaderboard(ctx, limit, offset)
	if err != nil {
		return nil, nil, 0, err
	}

	me, err := u.repo.LeaderboardEntryOf(ctx, userID)
	if err != nil {
		return nil, nil, 0, err
	}

	return entries, me, total, nil
}

// Opponents は「相手」フィルタの選択肢になる、対戦済みの相手一覧を返す。
func (u *StatsUsecase) Opponents(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	return u.repo.Opponents(ctx, userID)
}
