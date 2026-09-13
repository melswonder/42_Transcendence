package usecase

import (
	"context"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// Outcomes は集計対象の勝敗を古い順に返す（連勝の計算に使う）。
type StatsRepository interface {
	Summary(ctx context.Context, userID uuid.UUID, f MatchFilter) (domain.StatsSummary, error)
	Outcomes(ctx context.Context, userID uuid.UUID, f MatchFilter) ([]string, error)
	Timeseries(ctx context.Context, userID uuid.UUID, f MatchFilter, interval string) ([]domain.TimeseriesPoint, error)
	Breakdown(ctx context.Context, userID uuid.UUID, f MatchFilter) (domain.Breakdown, error)
	Leaderboard(ctx context.Context, limit, offset int) ([]domain.LeaderboardEntry, int, error)
	LeaderboardEntryOf(ctx context.Context, userID uuid.UUID) (*domain.LeaderboardEntry, error)
	Opponents(ctx context.Context, userID uuid.UUID) ([]domain.User, error)
}

type StatsUsecase struct {
	repo StatsRepository
}

func NewStatsUsecase(repo StatsRepository) *StatsUsecase {
	return &StatsUsecase{repo: repo}
}

// Summary の連勝は SQL で数えず、勝敗の列だけ取ってきて domain 側で数える。
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

// Leaderboard は自分が表示範囲の外にいても順位を出せるよう、me を別に引く。
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

// Opponents は「相手」フィルタの選択肢になる対戦済みの相手一覧。
func (u *StatsUsecase) Opponents(ctx context.Context, userID uuid.UUID) ([]domain.User, error) {
	return u.repo.Opponents(ctx, userID)
}
