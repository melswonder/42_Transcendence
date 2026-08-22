package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"transcendence-backend/domain"
)

// MediaRepo はアップロードのメタ情報の永続化。ファイル本体は FileStore が持つ。
type MediaRepo struct {
	db *gorm.DB
}

func NewMediaRepo(db *gorm.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

func (r *MediaRepo) CreateAsset(ctx context.Context, asset *domain.MediaAsset, storageKey string) error {
	row := MediaAsset{
		ID:               asset.ID,
		OwnerUserID:      asset.OwnerUserID,
		Purpose:          asset.Purpose,
		StorageKey:       storageKey,
		OriginalFilename: asset.OriginalFilename,
		MimeType:         asset.MimeType,
		SizeBytes:        asset.SizeBytes,
		Width:            asset.Width,
		Height:           asset.Height,
		ChecksumSHA256:   asset.ChecksumSHA256,
		Status:           asset.Status,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("insert media asset: %w", err)
	}
	asset.CreatedAt = row.CreatedAt
	return nil
}

func (r *MediaRepo) GetOwnedAsset(ctx context.Context, assetID, ownerID uuid.UUID) (*domain.MediaAsset, error) {
	var row MediaAsset
	err := r.db.WithContext(ctx).
		First(&row, "id = ? AND owner_user_id = ?", assetID, ownerID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrMediaNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find media asset: %w", err)
	}
	return toDomainMediaAsset(&row), nil
}

func (r *MediaRepo) GetActiveAsset(ctx context.Context, assetID uuid.UUID) (*domain.MediaAsset, string, error) {
	var row MediaAsset
	err := r.db.WithContext(ctx).
		First(&row, "id = ? AND status = ?", assetID, domain.MediaStatusActive).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", domain.ErrMediaNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("find media asset: %w", err)
	}
	return toDomainMediaAsset(&row), row.StorageKey, nil
}

func (r *MediaRepo) ListAssets(ctx context.Context, ownerID uuid.UUID, purpose string, limit, offset int) ([]domain.MediaAsset, int, error) {
	query := r.db.WithContext(ctx).Model(&MediaAsset{}).
		Where("owner_user_id = ? AND status = ?", ownerID, domain.MediaStatusActive)
	if purpose != "" {
		query = query.Where("purpose = ?", purpose)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count media assets: %w", err)
	}

	var rows []MediaAsset
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list media assets: %w", err)
	}

	assets := make([]domain.MediaAsset, 0, len(rows))
	for i := range rows {
		assets = append(assets, *toDomainMediaAsset(&rows[i]))
	}
	return assets, int(total), nil
}

// SoftDeleteAsset は論理削除。使用中のアバターなら users から外し、
// その場でデフォルトアバターに戻る。同一トランザクションで行う。
func (r *MediaRepo) SoftDeleteAsset(ctx context.Context, assetID, ownerID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&MediaAsset{}).
			Where("id = ? AND owner_user_id = ? AND status = ?", assetID, ownerID, domain.MediaStatusActive).
			Updates(map[string]any{
				"status":     domain.MediaStatusDeleted,
				"deleted_at": gorm.Expr("now()"),
			})
		if res.Error != nil {
			return fmt.Errorf("soft delete media asset: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return domain.ErrMediaNotFound
		}

		err := tx.Model(&User{}).
			Where("id = ? AND avatar_asset_id = ?", ownerID, assetID).
			Update("avatar_asset_id", nil).Error
		if err != nil {
			return fmt.Errorf("unset avatar: %w", err)
		}
		return nil
	})
}

func toDomainMediaAsset(m *MediaAsset) *domain.MediaAsset {
	return &domain.MediaAsset{
		ID:               m.ID,
		OwnerUserID:      m.OwnerUserID,
		Purpose:          m.Purpose,
		OriginalFilename: m.OriginalFilename,
		MimeType:         m.MimeType,
		SizeBytes:        m.SizeBytes,
		Width:            m.Width,
		Height:           m.Height,
		ChecksumSHA256:   m.ChecksumSHA256,
		Status:           m.Status,
		CreatedAt:        m.CreatedAt,
	}
}

// LocalFileStore はローカルディスクへの保存。
// 単一インスタンス構成なのでこれで足りる。S3 等に移すときはこの型だけ差し替える。
type LocalFileStore struct {
	baseDir string
}

func NewLocalFileStore(baseDir string) (*LocalFileStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}
	return &LocalFileStore{baseDir: baseDir}, nil
}

// resolve はキーを baseDir 配下の実パスへ解決する。
// ".." などで外へ出るキーは拒否する（パストラバーサル対策）。
func (s *LocalFileStore) resolve(key string) (string, error) {
	path := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if !strings.HasPrefix(path, filepath.Clean(s.baseDir)+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid storage key: %q", key)
	}
	return path, nil
}

func (s *LocalFileStore) Save(key string, data []byte) error {
	path, err := s.resolve(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (s *LocalFileStore) Open(key string) (io.ReadSeekCloser, error) {
	path, err := s.resolve(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}
