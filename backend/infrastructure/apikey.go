package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"transcendence-backend/domain"
)

// APIKeyRepo は api_keys の永続化。
type APIKeyRepo struct {
	db *gorm.DB
}

func NewAPIKeyRepo(db *gorm.DB) *APIKeyRepo {
	return &APIKeyRepo{db: db}
}

func (r *APIKeyRepo) CreateKey(ctx context.Context, key *domain.APIKey, keyHash string) error {
	row := APIKey{
		ID:        key.ID,
		UserID:    key.UserID,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		KeyHash:   keyHash,
		Scopes:    strings.Join(key.Scopes, ","),
		ExpiresAt: key.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("insert api key: %w", err)
	}
	key.CreatedAt = row.CreatedAt
	return nil
}

func (r *APIKeyRepo) ListKeys(ctx context.Context, userID uuid.UUID) ([]domain.APIKey, error) {
	var rows []APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	keys := make([]domain.APIKey, 0, len(rows))
	for i := range rows {
		keys = append(keys, *toDomainAPIKey(&rows[i]))
	}
	return keys, nil
}

func (r *APIKeyRepo) FindByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	var row APIKey
	err := r.db.WithContext(ctx).First(&row, "key_hash = ?", keyHash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find api key: %w", err)
	}
	return toDomainAPIKey(&row), nil
}

func (r *APIKeyRepo) RevokeKey(ctx context.Context, keyID, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&APIKey{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", keyID, userID).
		Update("revoked_at", gorm.Expr("now()"))
	if res.Error != nil {
		return fmt.Errorf("revoke api key: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrAPIKeyNotFound
	}
	return nil
}

// TouchKey は last_used_at を進める。sessions の touch と同じで、失敗は握りつぶす。
func (r *APIKeyRepo) TouchKey(ctx context.Context, keyID uuid.UUID) {
	_ = r.db.WithContext(ctx).Model(&APIKey{}).
		Where("id = ?", keyID).
		Update("last_used_at", gorm.Expr("now()")).Error
}

func toDomainAPIKey(k *APIKey) *domain.APIKey {
	return &domain.APIKey{
		ID:         k.ID,
		UserID:     k.UserID,
		Name:       k.Name,
		KeyPrefix:  k.KeyPrefix,
		Scopes:     strings.Split(k.Scopes, ","),
		ExpiresAt:  k.ExpiresAt,
		RevokedAt:  k.RevokedAt,
		LastUsedAt: k.LastUsedAt,
		CreatedAt:  k.CreatedAt,
	}
}

// FixedWindowLimiter はキー単位の固定窓レートリミッタ。
//
// 窓の開始から 1 分間のリクエスト数を数え、上限を超えたら 429。
// メモリ上なので再起動で消えるが、流量制限は揮発してよい情報。
// 複数インスタンス構成にするときは Redis に移す（README #3）。
type FixedWindowLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu      sync.Mutex
	buckets map[uuid.UUID]*windowBucket
}

type windowBucket struct {
	windowStart time.Time
	count       int
}

func NewFixedWindowLimiter(limitPerMinute int) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:   limitPerMinute,
		window:  time.Minute,
		now:     time.Now,
		buckets: make(map[uuid.UUID]*windowBucket),
	}
}

func (l *FixedWindowLimiter) Allow(keyID uuid.UUID) (bool, int, int, time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b, ok := l.buckets[keyID]
	if !ok || now.Sub(b.windowStart) >= l.window {
		b = &windowBucket{windowStart: now}
		l.buckets[keyID] = b
	}
	resetAt := b.windowStart.Add(l.window)

	if b.count >= l.limit {
		return false, l.limit, 0, resetAt
	}
	b.count++
	return true, l.limit, l.limit - b.count, resetAt
}
