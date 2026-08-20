package usecase

import (
	"context"
	"errors"
	"fmt"
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
