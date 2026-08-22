package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// fakeAuthRepo はメール登録に関わる部分だけを再現する。
type fakeAuthRepo struct {
	mu       sync.Mutex
	byEmail  map[string]*fakeAccount
	handles  map[string]bool
	sessions int
}

type fakeAccount struct {
	user domain.User
	hash *string
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{byEmail: make(map[string]*fakeAccount), handles: make(map[string]bool)}
}

func (r *fakeAuthRepo) FindUserByOAuth(context.Context, string, string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (r *fakeAuthRepo) CreateUserWithOAuth(context.Context, domain.OAuthProfile, string) (*domain.User, error) {
	return nil, errors.New("not used")
}

func (r *fakeAuthRepo) CreateUserWithPassword(
	_ context.Context, email, passwordHash, displayName, handle, locale string,
) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byEmail[email]; ok {
		return nil, domain.ErrEmailTaken
	}
	if r.handles[handle] {
		return nil, domain.ErrHandleTaken
	}
	user := domain.User{
		ID: uuid.New(), Email: &email, DisplayName: displayName,
		Handle: handle, PreferredLocale: locale, Status: "active",
	}
	r.byEmail[email] = &fakeAccount{user: user, hash: &passwordHash}
	r.handles[handle] = true
	return &user, nil
}

func (r *fakeAuthRepo) FindUserWithPasswordByEmail(_ context.Context, email string) (*domain.User, *string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.byEmail[email]
	if !ok {
		return nil, nil, domain.ErrUserNotFound
	}
	u := a.user
	return &u, a.hash, nil
}

func (r *fakeAuthRepo) CreateSession(context.Context, domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions++
	return nil
}

func (r *fakeAuthRepo) FindUserBySessionToken(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrSessionNotFound
}

func (r *fakeAuthRepo) RevokeSession(context.Context, string) error { return nil }

func validInput() RegisterInput {
	return RegisterInput{
		Email:       "alice@example.com",
		Password:    "correct horse battery staple",
		DisplayName: "Alice",
		Handle:      "alice_42",
	}
}

func TestRegisterAndLoginWithPassword(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	uc := NewAuthUsecase(nil, repo)
	ctx := context.Background()

	result, err := uc.Register(ctx, validInput())
	if err != nil {
		t.Fatalf("登録できるはず: %v", err)
	}
	if result.SessionToken == "" {
		t.Error("登録と同時にセッションが発行されるはず")
	}
	// パスワードは平文で保存されない。
	stored := *repo.byEmail["alice@example.com"].hash
	if stored == "correct horse battery staple" {
		t.Error("パスワードがハッシュ化されていない")
	}

	// 正しい資格情報でログインできる。
	if _, err := uc.LoginWithPassword(ctx, "alice@example.com", "correct horse battery staple"); err != nil {
		t.Errorf("ログインできるはず: %v", err)
	}
	// 誤ったパスワード・未知のメールはどちらも同じエラー。
	if _, err := uc.LoginWithPassword(ctx, "alice@example.com", "wrong password"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("誤パスワードは invalid_credentials: %v", err)
	}
	if _, err := uc.LoginWithPassword(ctx, "nobody@example.com", "whatever!"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("未知のメールも invalid_credentials: %v", err)
	}
}

func TestRegisterValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mutate  func(*RegisterInput)
		wantErr error
	}{
		{"メール形式が不正", func(in *RegisterInput) { in.Email = "not-an-email" }, domain.ErrInvalidEmail},
		{"パスワードが短い", func(in *RegisterInput) { in.Password = "short" }, domain.ErrWeakPassword},
		{"表示名が空", func(in *RegisterInput) { in.DisplayName = "  " }, domain.ErrInvalidDisplayName},
		{"handle が不正", func(in *RegisterInput) { in.Handle = "Bad Handle!" }, domain.ErrInvalidHandle},
		{"未対応ロケール", func(in *RegisterInput) { in.PreferredLocale = "de" }, domain.ErrInvalidLocale},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := NewAuthUsecase(nil, newFakeAuthRepo())
			in := validInput()
			tt.mutate(&in)
			if _, err := uc.Register(context.Background(), in); !errors.Is(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	uc := NewAuthUsecase(nil, repo)
	ctx := context.Background()

	if _, err := uc.Register(ctx, validInput()); err != nil {
		t.Fatal(err)
	}
	in := validInput()
	in.Handle = "alice_other"
	if _, err := uc.Register(ctx, in); !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("メール重複は email_taken: %v", err)
	}
}

// OAuth 専用アカウント（パスワード無し）にはパスワードでログインできない。
func TestLoginRejectsOAuthOnlyAccount(t *testing.T) {
	t.Parallel()

	repo := newFakeAuthRepo()
	email := "oauth@example.com"
	repo.byEmail[email] = &fakeAccount{
		user: domain.User{ID: uuid.New(), Email: &email, Handle: "oauth_user", Status: "active"},
		hash: nil,
	}
	uc := NewAuthUsecase(nil, repo)
	if _, err := uc.LoginWithPassword(context.Background(), email, "any password"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("OAuth 専用も invalid_credentials: %v", err)
	}
}
