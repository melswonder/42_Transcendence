package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"net/http"
	"slices"

	// アバターとして受け付ける形式のデコーダを image に登録する。
	_ "image/jpeg"
	_ "image/png"

	"github.com/google/uuid"

	// webp は標準ライブラリにデコーダが無いので x/image から登録する。
	_ "golang.org/x/image/webp"

	"transcendence-backend/domain"
)

// MediaRepository はアップロード済みファイルのメタ情報の永続化。
type MediaRepository interface {
	CreateAsset(ctx context.Context, asset *domain.MediaAsset, storageKey string) error
	// GetOwnedAsset は owner のアセットを返す。他人のものは domain.ErrMediaNotFound。
	GetOwnedAsset(ctx context.Context, assetID, ownerID uuid.UUID) (*domain.MediaAsset, error)
	// GetActiveAsset は配信用。所有者を問わず active なものだけ返す。
	GetActiveAsset(ctx context.Context, assetID uuid.UUID) (asset *domain.MediaAsset, storageKey string, err error)
	ListAssets(ctx context.Context, ownerID uuid.UUID, purpose string, limit, offset int) ([]domain.MediaAsset, int, error)
	// SoftDeleteAsset は論理削除し、アバターとして使われていれば外す（デフォルトに戻る）。
	SoftDeleteAsset(ctx context.Context, assetID, ownerID uuid.UUID) error
}

// FileStore はファイル本体の置き場所。実装はローカルディスク。
type FileStore interface {
	Save(key string, data []byte) error
	Open(key string) (io.ReadSeekCloser, error)
}

// MediaUsecase はアバター画像の受け入れ・配信・削除を進める。
type MediaUsecase struct {
	repo  MediaRepository
	store FileStore
}

func NewMediaUsecase(repo MediaRepository, store FileStore) *MediaUsecase {
	return &MediaUsecase{repo: repo, store: store}
}

// UploadAvatar は画像を検証してから保存する。
//
// 検証は申告された Content-Type ではなく中身で行う:
// 1. サイズ上限（handler 側の MaxBytesReader が第一関門、ここは保険）
// 2. 先頭バイトの sniff で MIME を判定し、許可リストと突き合わせる
// 3. 実際に画像としてデコードできることを確かめ、寸法を取る
func (u *MediaUsecase) UploadAvatar(ctx context.Context, ownerID uuid.UUID, filename string, data []byte) (*domain.MediaAsset, error) {
	if int64(len(data)) > domain.AvatarMaxBytes {
		return nil, domain.ErrMediaTooLarge
	}
	if len(data) == 0 {
		return nil, domain.ErrInvalidImage
	}

	mimeType := http.DetectContentType(data)
	if !slices.Contains(domain.AllowedAvatarMIMEs, mimeType) {
		return nil, domain.ErrUnsupportedMediaType
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, domain.ErrInvalidImage
	}
	// sniff とデコーダの見解が食い違うファイルは受け取らない（偽装対策）。
	if "image/"+format != mimeType {
		return nil, domain.ErrUnsupportedMediaType
	}

	sum := sha256.Sum256(data)
	storageKey := newStorageKey(format)
	if err := u.store.Save(storageKey, data); err != nil {
		return nil, fmt.Errorf("save avatar: %w", err)
	}

	asset := &domain.MediaAsset{
		ID:               uuid.New(),
		OwnerUserID:      ownerID,
		Purpose:          domain.MediaPurposeAvatar,
		OriginalFilename: trimFilename(filename),
		MimeType:         mimeType,
		SizeBytes:        int64(len(data)),
		Width:            &cfg.Width,
		Height:           &cfg.Height,
		ChecksumSHA256:   hex.EncodeToString(sum[:]),
		Status:           domain.MediaStatusActive,
	}
	if err := u.repo.CreateAsset(ctx, asset, storageKey); err != nil {
		return nil, err
	}
	return asset, nil
}

// OpenAsset は配信用にファイルを開く。active なものだけ。
func (u *MediaUsecase) OpenAsset(ctx context.Context, assetID uuid.UUID) (*domain.MediaAsset, io.ReadSeekCloser, error) {
	asset, storageKey, err := u.repo.GetActiveAsset(ctx, assetID)
	if err != nil {
		return nil, nil, err
	}
	file, err := u.store.Open(storageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("open asset %s: %w", assetID, err)
	}
	return asset, file, nil
}

// GetOwned は自分のアセット 1 件。
func (u *MediaUsecase) GetOwned(ctx context.Context, assetID, ownerID uuid.UUID) (*domain.MediaAsset, error) {
	return u.repo.GetOwnedAsset(ctx, assetID, ownerID)
}

// List は自分のアップロード一覧。
func (u *MediaUsecase) List(ctx context.Context, ownerID uuid.UUID, purpose string, limit, offset int) ([]domain.MediaAsset, int, error) {
	return u.repo.ListAssets(ctx, ownerID, purpose, limit, offset)
}

// Delete は論理削除。使用中のアバターならデフォルトに戻る（repo が同一トランザクションで外す）。
// ファイル本体は消さない。配信は status を見て止まるので急がなくてよい。
func (u *MediaUsecase) Delete(ctx context.Context, assetID, ownerID uuid.UUID) error {
	return u.repo.SoftDeleteAsset(ctx, assetID, ownerID)
}

// newStorageKey は推測できない保存名を作る。URL からの列挙を防ぐ。
func newStorageKey(format string) string {
	return fmt.Sprintf("avatars/%s.%s", rand.Text(), format)
}

// trimFilename は varchar(255) に収まるよう末尾を優先して詰める。
func trimFilename(name string) string {
	const maxLen = 255
	if len(name) <= maxLen {
		return name
	}
	return name[len(name)-maxLen:]
}
