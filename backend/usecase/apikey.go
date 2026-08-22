package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// APIKeyRepository は API キーの永続化。
type APIKeyRepository interface {
	CreateKey(ctx context.Context, key *domain.APIKey, keyHash string) error
	ListKeys(ctx context.Context, userID uuid.UUID) ([]domain.APIKey, error)
	// FindByHash は raw key のハッシュから引く。無ければ domain.ErrAPIKeyNotFound。
	FindByHash(ctx context.Context, keyHash string) (*domain.APIKey, error)
	// RevokeKey は自分のキーを失効させる。無ければ domain.ErrAPIKeyNotFound。
	RevokeKey(ctx context.Context, keyID, userID uuid.UUID) error
	// TouchKey は last_used_at を進める。失敗しても認証は成立している。
	TouchKey(ctx context.Context, keyID uuid.UUID)
}

// RateLimiter はキー単位の流量制限。実装はメモリ上の固定窓。
type RateLimiter interface {
	// Allow は 1 リクエストぶん消費を試みる。
	Allow(keyID uuid.UUID) (allowed bool, limit, remaining int, resetAt time.Time)
}

// APIKeyIdentity は認証済みリクエストの主体。handler が context で持ち回す。
type APIKeyIdentity struct {
	KeyID  uuid.UUID
	UserID uuid.UUID
	Scopes []string
}

// APIKeyUsecase はキーの発行・失効と、Public API の認証を進める。
type APIKeyUsecase struct {
	repo    APIKeyRepository
	limiter RateLimiter
	now     func() time.Time
}

func NewAPIKeyUsecase(repo APIKeyRepository, limiter RateLimiter) *APIKeyUsecase {
	return &APIKeyUsecase{repo: repo, limiter: limiter, now: time.Now}
}

// Create はキーを発行し、raw key を一度だけ返す。DB にはハッシュしか残らない。
func (u *APIKeyUsecase) Create(
	ctx context.Context, userID uuid.UUID, name string, scopes []string, expiresAt *time.Time,
) (raw string, key *domain.APIKey, err error) {
	if err := domain.ValidateAPIKeyInput(name, scopes, expiresAt, u.now()); err != nil {
		return "", nil, err
	}

	raw = domain.NewAPIKeySecret()
	key = &domain.APIKey{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		KeyPrefix: domain.APIKeyPrefixOf(raw),
		Scopes:    scopes,
		ExpiresAt: expiresAt,
	}
	if err := u.repo.CreateKey(ctx, key, domain.HashAPIKey(raw)); err != nil {
		return "", nil, err
	}
	return raw, key, nil
}

// List は自分のキー一覧。raw key は含まれない（もうどこにも無い）。
func (u *APIKeyUsecase) List(ctx context.Context, userID uuid.UUID) ([]domain.APIKey, error) {
	return u.repo.ListKeys(ctx, userID)
}

// Revoke はキーを失効させる。以後そのキーのリクエストは拒否される。
func (u *APIKeyUsecase) Revoke(ctx context.Context, keyID, userID uuid.UUID) error {
	return u.repo.RevokeKey(ctx, keyID, userID)
}

// Authenticate は raw key を検証して主体を返す。
// 失効・期限切れは 401 相当の別々のエラーで返し、原因をクライアントに伝える。
func (u *APIKeyUsecase) Authenticate(ctx context.Context, raw string) (*APIKeyIdentity, error) {
	if raw == "" {
		return nil, domain.ErrAPIKeyInvalid
	}
	key, err := u.repo.FindByHash(ctx, domain.HashAPIKey(raw))
	if err != nil {
		return nil, domain.ErrAPIKeyInvalid // 存在の有無は区別せず invalid に丸める
	}
	if key.Revoked() {
		return nil, domain.ErrAPIKeyRevoked
	}
	if key.Expired(u.now()) {
		return nil, domain.ErrAPIKeyExpired
	}

	u.repo.TouchKey(ctx, key.ID)
	return &APIKeyIdentity{KeyID: key.ID, UserID: key.UserID, Scopes: key.Scopes}, nil
}

// Authorize はスコープと流量を確かめる。認証の後、ハンドラ本体の前に呼ぶ。
func (u *APIKeyUsecase) Authorize(identity *APIKeyIdentity, scope string) (limit, remaining int, resetAt time.Time, err error) {
	key := domain.APIKey{Scopes: identity.Scopes}
	if !key.HasScope(scope) {
		return 0, 0, time.Time{}, domain.ErrAPIKeyScope
	}
	allowed, limit, remaining, resetAt := u.limiter.Allow(identity.KeyID)
	if !allowed {
		return limit, remaining, resetAt, domain.ErrAPIKeyRateLimited
	}
	return limit, remaining, resetAt, nil
}
