package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"transcendence-backend/usecase"
)

// AchievementRepo は解除済み実績の永続化。定義そのものは domain 側にある。
type AchievementRepo struct {
	db *gorm.DB
}

var _ usecase.AchievementRepository = (*AchievementRepo)(nil)

func NewAchievementRepo(db *gorm.DB) *AchievementRepo {
	return &AchievementRepo{db: db}
}

// UnlockedAt は code → 解除時刻を返す。
func (r *AchievementRepo) UnlockedAt(ctx context.Context, userID uuid.UUID) (map[string]time.Time, error) {
	var rows []UserAchievement

	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("select user achievements: %w", err)
	}

	unlocked := make(map[string]time.Time, len(rows))
	for _, row := range rows {
		unlocked[row.Code] = row.UnlockedAt
	}

	return unlocked, nil
}

// Unlock は解除済みとして記録する。
//
// 同じ実績を二重に解除しようとしても落とさない（DoNothing）。
// 対戦の記録と同時に走るので、競合しても静かに片方が勝てばよい。
func (r *AchievementRepo) Unlock(ctx context.Context, userID uuid.UUID, codes []string) error {
	if len(codes) == 0 {
		return nil
	}

	rows := make([]UserAchievement, 0, len(codes))
	for _, code := range codes {
		rows = append(rows, UserAchievement{UserID: userID, Code: code, UnlockedAt: time.Now()})
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&rows).Error
	if err != nil {
		return fmt.Errorf("insert user achievements: %w", err)
	}

	return nil
}
