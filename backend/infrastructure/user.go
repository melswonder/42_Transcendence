package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// UserRepo はプロフィールとユーザー検索の永続化。
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetProfile は自分のプロフィール。連携プロバイダとパスワード有無も添える。
func (r *UserRepo) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, "id = ? AND anonymized_at IS NULL", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	var providers []string
	err = r.db.WithContext(ctx).Model(&OAuthAccount{}).
		Where("user_id = ?", userID).
		Order("provider").
		Pluck("provider", &providers).Error
	if err != nil {
		return nil, fmt.Errorf("list oauth providers: %w", err)
	}

	return &domain.Profile{
		User:            *toDomainUser(&user),
		HasPassword:     user.PasswordHash != nil,
		LinkedProviders: providers,
	}, nil
}

// GetUser は公開プロフィール用。退会済みは見つからない扱い。
func (r *UserRepo) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, "id = ? AND anonymized_at IS NULL", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}
	return toDomainUser(&user), nil
}

// UpdateProfile は指定されたフィールドだけを更新する。
func (r *UserRepo) UpdateProfile(ctx context.Context, userID uuid.UUID, update domain.ProfileUpdate) (*domain.Profile, error) {
	values := map[string]any{}
	if update.DisplayName != nil {
		values["display_name"] = *update.DisplayName
	}
	if update.Handle != nil {
		values["handle"] = *update.Handle
	}
	if update.PreferredLocale != nil {
		values["preferred_locale"] = *update.PreferredLocale
	}
	if update.AvatarSet {
		values["avatar_asset_id"] = update.AvatarAssetID // nil なら外れてデフォルトに戻る
	}

	if len(values) > 0 {
		err := r.db.WithContext(ctx).Model(&User{}).
			Where("id = ?", userID).
			Updates(values).Error
		if err != nil {
			return nil, translateUniqueViolation(err)
		}
	}
	return r.GetProfile(ctx, userID)
}

// SearchUsers は viewer から見えるユーザーを探す。
// 自分自身・退会済み・どちらか一方でもブロックしている相手は除く。
func (r *UserRepo) SearchUsers(ctx context.Context, viewerID uuid.UUID, f usecase.UserSearchFilter) ([]domain.User, int, error) {
	query := r.db.WithContext(ctx).Model(&User{}).
		Where("id <> ?", viewerID).
		Where("anonymized_at IS NULL AND status = 'active'").
		Where(`NOT EXISTS (
			SELECT 1 FROM blocks b
			WHERE (b.blocker_user_id = ? AND b.blocked_user_id = users.id)
			   OR (b.blocker_user_id = users.id AND b.blocked_user_id = ?)
		)`, viewerID, viewerID)

	if f.Handle != "" {
		query = query.Where("handle = ?", f.Handle)
	}
	if f.Query != "" {
		pattern := "%" + f.Query + "%"
		query = query.Where("display_name ILIKE ? OR handle ILIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	var rows []User
	err := query.Order("handle").Limit(f.Limit).Offset(f.Offset).Find(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("search users: %w", err)
	}

	users := make([]domain.User, 0, len(rows))
	for i := range rows {
		users = append(users, *toDomainUser(&rows[i]))
	}
	return users, int(total), nil
}

// GetAvatarAsset は PATCH /users/me の avatar_asset_id 検証用。
func (r *UserRepo) GetAvatarAsset(ctx context.Context, assetID, ownerID uuid.UUID) (*domain.MediaAsset, error) {
	var asset MediaAsset
	err := r.db.WithContext(ctx).
		First(&asset, "id = ? AND owner_user_id = ?", assetID, ownerID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMediaNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find media asset: %w", err)
	}
	return toDomainMediaAsset(&asset), nil
}
