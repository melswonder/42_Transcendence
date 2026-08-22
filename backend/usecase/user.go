package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// UserSearchFilter はユーザー検索の絞り込み。
type UserSearchFilter struct {
	Query  string // display_name / handle の部分一致
	Handle string // handle の完全一致
	Limit  int
	Offset int
}

// UserRepository はユーザーの読み書き。実装は infrastructure 層にある。
type UserRepository interface {
	// GetProfile は自分のプロフィール（連携プロバイダ含む）を返す。
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	// GetUser は公開プロフィール用の 1 人。退会済みは domain.ErrUserNotFound。
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	// UpdateProfile は nil でないフィールドだけを更新して結果を返す。
	// handle の衝突は domain.ErrHandleTaken。
	UpdateProfile(ctx context.Context, userID uuid.UUID, update domain.ProfileUpdate) (*domain.Profile, error)
	// SearchUsers は viewer から見えるユーザーを返す。
	// 退会済みと、どちらかがブロックしている相手は除く。
	SearchUsers(ctx context.Context, viewerID uuid.UUID, f UserSearchFilter) ([]domain.User, int, error)
	// GetAvatarAsset はアバター指定の検証用。owner のものでなければ見つからない扱い。
	GetAvatarAsset(ctx context.Context, assetID, ownerID uuid.UUID) (*domain.MediaAsset, error)
}

// UserUsecase はプロフィールの表示・編集とユーザー検索を進める。
type UserUsecase struct {
	repo UserRepository
}

func NewUserUsecase(repo UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

// MyProfile は自分の非公開プロフィールを返す。
func (u *UserUsecase) MyProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	return u.repo.GetProfile(ctx, userID)
}

// UpdateMe はプロフィールを編集する。アバターの指定は
// 「自分のアクティブなアバター用アセット」であることを確かめてから通す。
func (u *UserUsecase) UpdateMe(ctx context.Context, userID uuid.UUID, update domain.ProfileUpdate) (*domain.Profile, error) {
	if err := update.Validate(); err != nil {
		return nil, err
	}

	if update.AvatarSet && update.AvatarAssetID != nil {
		asset, err := u.repo.GetAvatarAsset(ctx, *update.AvatarAssetID, userID)
		if err != nil {
			if errors.Is(err, domain.ErrMediaNotFound) {
				return nil, domain.ErrInvalidAvatarAsset
			}
			return nil, err
		}
		if asset.Purpose != domain.MediaPurposeAvatar || asset.Status != domain.MediaStatusActive {
			return nil, domain.ErrInvalidAvatarAsset
		}
	}

	return u.repo.UpdateProfile(ctx, userID, update)
}

// GetPublic は他人から見えるプロフィールを返す。
func (u *UserUsecase) GetPublic(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return u.repo.GetUser(ctx, userID)
}

// Search はフレンド申請の相手探しなどに使うユーザー検索。
func (u *UserUsecase) Search(ctx context.Context, viewerID uuid.UUID, f UserSearchFilter) ([]domain.User, int, error) {
	return u.repo.SearchUsers(ctx, viewerID, f)
}
