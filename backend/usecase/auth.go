package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"transcendence-backend/domain"
)

// OAuthProvider の AuthCodeURL は同意画面の URL、ExchangeCode は認可コードから検証済みのプロフィールを返す。
type OAuthProvider interface {
	AuthCodeURL(state, nonce string) string
	ExchangeCode(ctx context.Context, code, nonce string) (domain.OAuthProfile, error)
}

// FindUserByOAuth は (provider, sub) で引き、いなければ domain.ErrUserNotFound。
// CreateUserWithOAuth はユーザーと OAuth 連携をまとめて作る。handle が埋まっていれば domain.ErrHandleTaken。
// FindUserBySessionToken は期限切れ・失効済みなら domain.ErrSessionNotFound。
// RevokeSession は対象が無くてもエラーにしない。
type AuthRepository interface {
	FindUserByOAuth(ctx context.Context, provider, providerAccountID string) (*domain.User, error)
	CreateUserWithOAuth(ctx context.Context, profile domain.OAuthProfile, handle string) (*domain.User, error)
	// CreateUserWithPassword はメール重複で domain.ErrEmailTaken、handle 重複で domain.ErrHandleTaken。
	CreateUserWithPassword(ctx context.Context, email, passwordHash, displayName, handle, locale string) (*domain.User, error)
	// FindUserWithPasswordByEmail は OAuth 専用ユーザーならハッシュが nil。無ければ domain.ErrUserNotFound。
	FindUserWithPasswordByEmail(ctx context.Context, email string) (*domain.User, *string, error)
	CreateSession(ctx context.Context, session domain.Session) error
	FindUserBySessionToken(ctx context.Context, tokenHash string) (*domain.User, error)
	RevokeSession(ctx context.Context, tokenHash string) error
}

// LoginStart の State と Nonce は callback で照合するので、呼び出し側が Cookie などに預かる。
type LoginStart struct {
	AuthURL string
	State   string
	Nonce   string
}

// LoginResult の SessionToken は生トークンで、ここでしか手に入らない（DB にはハッシュしかない）。
type LoginResult struct {
	User         *domain.User
	SessionToken string
	ExpiresAt    time.Time
}

type AuthUsecase struct {
	provider OAuthProvider
	repo     AuthRepository
	now      func() time.Time // テストで時刻を差し替えられるようにしておく
}

func NewAuthUsecase(provider OAuthProvider, repo AuthRepository) *AuthUsecase {
	return &AuthUsecase{
		provider: provider,
		repo:     repo,
		now:      time.Now,
	}
}

func (u *AuthUsecase) StartLogin() LoginStart {
	state, nonce := domain.NewOAuthState()

	return LoginStart{
		AuthURL: u.provider.AuthCodeURL(state, nonce),
		State:   state,
		Nonce:   nonce,
	}
}

// CompleteLogin は callback で受け取った認可コードをプロフィールに交換し、
// ユーザーを探す（いなければ作る）→ セッションを発行する。
func (u *AuthUsecase) CompleteLogin(ctx context.Context, code, nonce string) (LoginResult, error) {
	profile, err := u.provider.ExchangeCode(ctx, code, nonce)
	if err != nil {
		return LoginResult{}, fmt.Errorf("exchange code: %w", err)
	}

	user, err := u.findOrCreateUser(ctx, profile)
	if err != nil {
		return LoginResult{}, err
	}

	return u.issueSession(ctx, user)
}

// issueSession は OAuth とメール認証で共用する。
func (u *AuthUsecase) issueSession(ctx context.Context, user *domain.User) (LoginResult, error) {
	rawToken := domain.NewSessionToken()
	session := domain.NewSession(user.ID, rawToken, u.now())
	if err := u.repo.CreateSession(ctx, session); err != nil {
		return LoginResult{}, fmt.Errorf("create session: %w", err)
	}

	return LoginResult{
		User:         user,
		SessionToken: rawToken,
		ExpiresAt:    session.ExpiresAt,
	}, nil
}

type RegisterInput struct {
	Email           string
	Password        string
	DisplayName     string
	Handle          string
	PreferredLocale string
}

// Register はメール+パスワードでアカウントを作り、そのままログイン状態にする。
func (u *AuthUsecase) Register(ctx context.Context, in RegisterInput) (LoginResult, error) {
	if err := domain.ValidateEmail(in.Email); err != nil {
		return LoginResult{}, err
	}
	if err := domain.ValidatePassword(in.Password); err != nil {
		return LoginResult{}, err
	}
	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" || len([]rune(displayName)) > domain.DisplayNameMaxLen {
		return LoginResult{}, domain.ErrInvalidDisplayName
	}
	if err := domain.ValidateHandle(in.Handle); err != nil {
		return LoginResult{}, err
	}
	locale := in.PreferredLocale
	if locale == "" {
		locale = "ja"
	}
	if !slices.Contains(domain.SupportedLocales, locale) {
		return LoginResult{}, domain.ErrInvalidLocale
	}

	hash, err := domain.HashPassword(in.Password)
	if err != nil {
		return LoginResult{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := u.repo.CreateUserWithPassword(ctx, in.Email, hash, displayName, in.Handle, locale)
	if err != nil {
		return LoginResult{}, err
	}
	return u.issueSession(ctx, user)
}

// LoginWithPassword はユーザーの有無・OAuth 専用・パスワード不一致をすべて同じエラーに丸め、
// ダミー照合で応答時間も揃える（アカウントの存在を悟らせない）。
func (u *AuthUsecase) LoginWithPassword(ctx context.Context, email, password string) (LoginResult, error) {
	user, hash, err := u.repo.FindUserWithPasswordByEmail(ctx, email)
	if err != nil || hash == nil {
		domain.CheckPasswordDummy()
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	if !domain.CheckPassword(*hash, password) {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	return u.issueSession(ctx, user)
}

func (u *AuthUsecase) Authenticate(ctx context.Context, rawToken string) (*domain.User, error) {
	if rawToken == "" {
		return nil, domain.ErrSessionNotFound
	}

	return u.repo.FindUserBySessionToken(ctx, domain.HashSessionToken(rawToken))
}

// Logout は Cookie を消さない（handler の仕事）。
func (u *AuthUsecase) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	return u.repo.RevokeSession(ctx, domain.HashSessionToken(rawToken))
}

// findOrCreateUser がメールではなく sub で探すのは、メールは Google 側で変わりうるから。
func (u *AuthUsecase) findOrCreateUser(ctx context.Context, profile domain.OAuthProfile) (*domain.User, error) {
	user, err := u.repo.FindUserByOAuth(ctx, profile.Provider, profile.ProviderAccountID)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find user by oauth: %w", err)
	}

	return u.createUser(ctx, profile)
}

// createUser は Google が handle をくれないのでこちらで組み立てる。
// 他人と衝突したら候補を変えて作り直す。
func (u *AuthUsecase) createUser(ctx context.Context, profile domain.OAuthProfile) (*domain.User, error) {
	base := domain.HandleBase(profile)

	for attempt := range domain.HandleGenerateAttempts {
		user, err := u.repo.CreateUserWithOAuth(ctx, profile, domain.HandleCandidate(base, attempt))
		switch {
		case err == nil:
			return user, nil
		case errors.Is(err, domain.ErrHandleTaken):
			continue
		default:
			return nil, fmt.Errorf("create user with oauth: %w", err)
		}
	}

	return nil, domain.ErrHandleUnavailable
}
