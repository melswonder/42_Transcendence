package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// UnlockedAt は code → 解除時刻を返す。
// Unlock は既に記録済みのものを無視する。
type AchievementRepository interface {
	UnlockedAt(ctx context.Context, userID uuid.UUID) (map[string]time.Time, error)
	Unlock(ctx context.Context, userID uuid.UUID, codes []string) error
}

type AchievementUsecase struct {
	repo  AchievementRepository
	stats *StatsUsecase
}

func NewAchievementUsecase(repo AchievementRepository, stats *StatsUsecase) *AchievementUsecase {
	return &AchievementUsecase{repo: repo, stats: stats}
}

// List は全実績を進捗つきで返す。未解除のものも含める。
func (u *AchievementUsecase) List(ctx context.Context, userID uuid.UUID) ([]domain.Achievement, error) {
	summary, err := u.stats.Summary(ctx, userID, MatchFilter{})
	if err != nil {
		return nil, err
	}

	unlocked, err := u.repo.UnlockedAt(ctx, userID)
	if err != nil {
		return nil, err
	}

	return domain.EvaluateAchievements(summary, unlocked), nil
}

// SyncAfterMatch を対戦の記録と同じトランザクションに入れないのは、
// 実績の書き込みに失敗しても対戦を取り消したくないため。
// 次の対戦のときに改めて判定されるので、取りこぼしても自然に追いつく。
func (u *AchievementUsecase) SyncAfterMatch(ctx context.Context, userID uuid.UUID) error {
	summary, err := u.stats.Summary(ctx, userID, MatchFilter{})
	if err != nil {
		return err
	}

	unlocked, err := u.repo.UnlockedAt(ctx, userID)
	if err != nil {
		return err
	}

	return u.repo.Unlock(ctx, userID, domain.NewlyUnlocked(summary, unlocked))
}
