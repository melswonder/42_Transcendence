package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// fakeAPIKeyRepo はメモリ上のキー置き場。ハッシュで引く挙動も再現する。
type fakeAPIKeyRepo struct {
	mu     sync.Mutex
	byHash map[string]*domain.APIKey
}

func newFakeAPIKeyRepo() *fakeAPIKeyRepo {
	return &fakeAPIKeyRepo{byHash: make(map[string]*domain.APIKey)}
}

func (r *fakeAPIKeyRepo) CreateKey(_ context.Context, key *domain.APIKey, keyHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copied := *key
	r.byHash[keyHash] = &copied
	return nil
}

func (r *fakeAPIKeyRepo) ListKeys(_ context.Context, userID uuid.UUID) ([]domain.APIKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.APIKey
	for _, k := range r.byHash {
		if k.UserID == userID {
			out = append(out, *k)
		}
	}
	return out, nil
}

func (r *fakeAPIKeyRepo) FindByHash(_ context.Context, keyHash string) (*domain.APIKey, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	k, ok := r.byHash[keyHash]
	if !ok {
		return nil, domain.ErrAPIKeyNotFound
	}
	copied := *k
	return &copied, nil
}

func (r *fakeAPIKeyRepo) RevokeKey(_ context.Context, keyID, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for _, k := range r.byHash {
		if k.ID == keyID && k.UserID == userID && k.RevokedAt == nil {
			k.RevokedAt = &now
			return nil
		}
	}
	return domain.ErrAPIKeyNotFound
}

func (r *fakeAPIKeyRepo) TouchKey(_ context.Context, _ uuid.UUID) {}

// allowAllLimiter はテスト用。常に許可する。
type allowAllLimiter struct{}

func (allowAllLimiter) Allow(uuid.UUID) (bool, int, int, time.Time) {
	return true, 60, 59, time.Now().Add(time.Minute)
}

func TestAPIKeyCreateAndAuthenticate(t *testing.T) {
	t.Parallel()

	repo := newFakeAPIKeyRepo()
	uc := NewAPIKeyUsecase(repo, allowAllLimiter{})
	userID := uuid.New()
	ctx := context.Background()

	raw, key, err := uc.Create(ctx, userID, "test key", []string{domain.APIScopeRead}, nil)
	if err != nil {
		t.Fatalf("発行できるはず: %v", err)
	}
	if raw == "" || key.KeyPrefix == "" {
		t.Fatal("raw key と prefix が返るはず")
	}
	// raw はどこにも保存されない。ハッシュだけが引ける。
	if _, ok := repo.byHash[raw]; ok {
		t.Error("raw key がそのまま保存されている")
	}
	if _, ok := repo.byHash[domain.HashAPIKey(raw)]; !ok {
		t.Error("ハッシュで保存されるはず")
	}

	identity, err := uc.Authenticate(ctx, raw)
	if err != nil || identity.UserID != userID {
		t.Fatalf("正しいキーは通るはず: %v %+v", err, identity)
	}

	// でたらめなキーは invalid。存在の有無は区別しない。
	if _, err := uc.Authenticate(ctx, "tsc_nonexistent"); !errors.Is(err, domain.ErrAPIKeyInvalid) {
		t.Errorf("不明なキーは invalid のはず: %v", err)
	}
}

func TestAPIKeyRejectsRevokedAndExpired(t *testing.T) {
	t.Parallel()

	repo := newFakeAPIKeyRepo()
	uc := NewAPIKeyUsecase(repo, allowAllLimiter{})
	userID := uuid.New()
	ctx := context.Background()

	// 失効させたキーは拒否。
	raw, key, err := uc.Create(ctx, userID, "to revoke", []string{domain.APIScopeRead}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := uc.Revoke(ctx, key.ID, userID); err != nil {
		t.Fatalf("失効できるはず: %v", err)
	}
	if _, err := uc.Authenticate(ctx, raw); !errors.Is(err, domain.ErrAPIKeyRevoked) {
		t.Errorf("失効済みは revoked のはず: %v", err)
	}
	// 他人は失効させられない。
	if err := uc.Revoke(ctx, key.ID, uuid.New()); !errors.Is(err, domain.ErrAPIKeyNotFound) {
		t.Errorf("他人のキーは失効できないはず: %v", err)
	}

	// 期限切れのキーは拒否。時計を進めて再現する。
	expiry := time.Now().Add(time.Hour)
	raw2, _, err := uc.Create(ctx, userID, "short lived", []string{domain.APIScopeRead}, &expiry)
	if err != nil {
		t.Fatal(err)
	}
	uc.now = func() time.Time { return time.Now().Add(2 * time.Hour) }
	if _, err := uc.Authenticate(ctx, raw2); !errors.Is(err, domain.ErrAPIKeyExpired) {
		t.Errorf("期限切れは expired のはず: %v", err)
	}
}

func TestAPIKeyScopeAndRateLimit(t *testing.T) {
	t.Parallel()

	// read だけのキーで write を要求するとスコープ不足。
	uc := NewAPIKeyUsecase(newFakeAPIKeyRepo(), allowAllLimiter{})
	identity := &APIKeyIdentity{KeyID: uuid.New(), Scopes: []string{domain.APIScopeRead}}
	if _, _, _, err := uc.Authorize(identity, domain.APIScopeWrite); !errors.Is(err, domain.ErrAPIKeyScope) {
		t.Errorf("スコープ不足は拒否されるはず: %v", err)
	}
	if _, _, _, err := uc.Authorize(identity, domain.APIScopeRead); err != nil {
		t.Errorf("持っているスコープは通るはず: %v", err)
	}
}

func TestAPIKeyInputValidation(t *testing.T) {
	t.Parallel()

	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name      string
		keyName   string
		scopes    []string
		expiresAt *time.Time
		wantErr   error
	}{
		{"有効な入力", "my key", []string{"read", "write"}, &future, nil},
		{"無期限も有効", "my key", []string{"read"}, nil, nil},
		{"名前が空", "  ", []string{"read"}, nil, domain.ErrInvalidAPIKeyInput},
		{"スコープが空", "k", nil, nil, domain.ErrInvalidAPIKeyInput},
		{"不明なスコープ", "k", []string{"admin"}, nil, domain.ErrInvalidAPIKeyInput},
		{"過去の期限", "k", []string{"read"}, &past, domain.ErrInvalidAPIKeyInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := domain.ValidateAPIKeyInput(tt.keyName, tt.scopes, tt.expiresAt, now)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}
