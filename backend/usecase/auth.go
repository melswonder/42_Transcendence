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

// OAuthProvider は外部プロバイダとの往復を抽象化する。
// AuthCodeURL はユーザーを飛ばす先（プロバイダの同意画面）の URL を組み立てる
// ExchangeCode は認可コードを、検証済みのプロフィールに変換する。
type OAuthProvider interface {
	AuthCodeURL(state, nonce string) string
	ExchangeCode(ctx context.Context, code, nonce string) (domain.OAuthProfile, error)
}

// AuthRepository は認証まわりの永続化。実装は infrastructure 層にある。
// FindUserByOAuth は (provider, sub) からユーザーを引く。いなければ domain.ErrUserNotFound。
// CreateUserWithOAuth はユーザーと OAuth 連携をまとめて作る。handle が埋まっていれば domain.ErrHandleTaken。
// CreateSession は発行済みのセッションを保存する。
// FindUserBySessionToken はトークンのハッシュから有効なセッションのユーザーを引く。期限切れ・失効済みは domain.ErrSessionNotFound。
// RevokeSession はセッションを失効させる。対象が無くてもエラーにしない。
type AuthRepository interface {
	FindUserByOAuth(ctx context.Context, provider, providerAccountID string) (*domain.User, error)
	CreateUserWithOAuth(ctx context.Context, profile domain.OAuthProfile, handle string) (*domain.User, error)
	// CreateUserWithPassword はメール登録のユーザーを作る。
	// メール重複は domain.ErrEmailTaken、handle 重複は domain.ErrHandleTaken。
	CreateUserWithPassword(ctx context.Context, email, passwordHash, displayName, handle, locale string) (*domain.User, error)
	// FindUserWithPasswordByEmail はメールからユーザーとパスワードハッシュを引く。
	// OAuth 専用ユーザーはハッシュが nil で返る。無ければ domain.ErrUserNotFound。
	FindUserWithPasswordByEmail(ctx context.Context, email string) (*domain.User, *string, error)
	CreateSession(ctx context.Context, session domain.Session) error
	FindUserBySessionToken(ctx context.Context, tokenHash string) (*domain.User, error)
	RevokeSession(ctx context.Context, tokenHash string) error
}

// LoginStart はログインを始めるのに必要なもの一式。
// state と nonce は callback で照合するので、呼び出し側が Cookie などに預かる。
type LoginStart struct {
	AuthURL string
	State   string
	Nonce   string
}

// LoginResult はログイン成立後の結果。
// SessionToken は生トークンで、ここでしか手に入らない（DB にはハッシュしかない）。
type LoginResult struct {
	User         *domain.User
	SessionToken string
	ExpiresAt    time.Time
}

// AuthUsecase はログインからログアウトまでの流れを進める。外を触るのは provider と repo の 2 つだけ。
// StartLogin はログインを始める。同意画面の URL と、照合用の state / nonce を返す。
// CompleteLogin は callback で受け取った認可コードを、ログイン済みの状態に変える。
// Authenticate は Cookie の生トークンから、ログイン中のユーザーを引く。
// Logout はセッションを失効させる。Cookie を消すのは handler の仕事。
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

// StartLogin はログインの 1 歩目。まだ誰が来たかは分からない。
func (u *AuthUsecase) StartLogin() LoginStart {
	state, nonce := domain.NewOAuthState()

	return LoginStart{
		AuthURL: u.provider.AuthCodeURL(state, nonce),
		State:   state,
		Nonce:   nonce,
	}
}

// CompleteLogin は callback で受け取った認可コードを、ログイン済みの状態に変える。
//
//  1. code をプロフィールに交換する（ここで初めて誰か分かる）
//  2. 既存ユーザーを探す。いなければ作る
//  3. セッションを発行する
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

// issueSession はログイン成立後のセッション発行。OAuth とメール認証で共用する。
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

// RegisterInput はメール+パスワードでの新規登録の入力。
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

// LoginWithPassword はメール+パスワードでのログイン。
// ユーザーの有無・OAuth 専用・パスワード不一致はすべて同じエラーに丸め、
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

// Authenticate は Cookie の生トークンからユーザーを引く。認証ミドルウェア用。
func (u *AuthUsecase) Authenticate(ctx context.Context, rawToken string) (*domain.User, error) {
	if rawToken == "" {
		return nil, domain.ErrSessionNotFound
	}

	return u.repo.FindUserBySessionToken(ctx, domain.HashSessionToken(rawToken))
}

// Logout はセッションを失効させる。Cookie を消すのは handler の仕事。
func (u *AuthUsecase) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	return u.repo.RevokeSession(ctx, domain.HashSessionToken(rawToken))
}

// findOrCreateUser は (provider, sub) でユーザーを引き、いなければ作る。
//
// 探すキーがメールではなく sub なのは、メールは Google 側で変わりうるから。
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

// createUser は handle を生成しながら新規ユーザーを作る。
//
// Google は handle をくれないのでこちらで組み立てるが、他人と衝突しうる。
// 衝突したら候補を変えて作り直す。
func (u *AuthUsecase) createUser(ctx context.Context, profile domain.OAuthProfile) (*domain.User, error) {
	base := domain.HandleBase(profile)

	for attempt := range domain.HandleGenerateAttempts {
		user, err := u.repo.CreateUserWithOAuth(ctx, profile, domain.HandleCandidate(base, attempt))
		switch {
		case err == nil:
			return user, nil
		case errors.Is(err, domain.ErrHandleTaken):
			continue // 別の候補で作り直す
		default:
			return nil, fmt.Errorf("create user with oauth: %w", err)
		}
	}

	return nil, domain.ErrHandleUnavailable
}
