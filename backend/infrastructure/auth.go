// 認証まわりの実装。Google とのやり取りと、DB への永続化の両方をここに置く。
// usecase.OAuthProvider と usecase.AuthRepository の実装がある。
package infrastructure

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// ----- Google OAuth 2.0 / OpenID Connect の具体的な手順 -----
// ここだけが oauth2 / idtoken ライブラリと Google の仕様を知っている。

// GoogleOAuthConfig は Google Cloud Console で発行した値。
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// GoogleOAuth は usecase.OAuthProvider の Google 実装。
// AuthCodeURL は accounts.google.com の同意画面 URL を組み立てる。
// ExchangeCode は認可コードをトークンに交換し、id_token を検証してプロフィールに詰め替える。
type GoogleOAuth struct {
	cfg      *oauth2.Config
	clientID string
}

var _ usecase.OAuthProvider = (*GoogleOAuth)(nil)

func NewGoogleOAuth(c GoogleOAuthConfig) *GoogleOAuth {
	return &GoogleOAuth{
		cfg: &oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			RedirectURL:  c.RedirectURL,
			// openid が無いと id_token が返らない。email / profile で claim が増える。
			Scopes:   []string{"openid", "email", "profile"},
			Endpoint: google.Endpoint,
		},
		clientID: c.ClientID,
	}
}

// AuthCodeURL は accounts.google.com の同意画面 URL を組み立てる。
// この画面は Google のものなので、こちらで作ることはない。
func (g *GoogleOAuth) AuthCodeURL(state, nonce string) string {
	return g.cfg.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		// refresh_token は使わないので offline にしない。
		oauth2.AccessTypeOnline,
		// 常にアカウント選択を出す。複数アカウント持ちが切り替えられるように。
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
}

// ExchangeCode は認可コードを検証済みプロフィールに変換する。
//
//  1. code をトークンに交換する（サーバー間通信。client_secret を使うのでここだけ安全に行える）
//  2. id_token の署名・iss・aud・exp を検証する
//  3. nonce が自分の出した値と一致するか確かめる
//  4. claim を domain.OAuthProfile に詰め替える
func (g *GoogleOAuth) ExchangeCode(ctx context.Context, code, nonce string) (domain.OAuthProfile, error) {
	token, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return domain.OAuthProfile{}, fmt.Errorf("exchange authorization code: %w", err)
	}

	// id_token は oauth2.Token の構造体フィールドではなく、生 JSON の追加項目として入っている。
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return domain.OAuthProfile{}, errors.New("google response has no id_token (openid scope missing?)")
	}

	// 署名検証。JWKS の取得とキャッシュはライブラリ任せ。自前で JWT をパースしない。
	payload, err := idtoken.Validate(ctx, rawIDToken, g.clientID)
	if err != nil {
		return domain.OAuthProfile{}, fmt.Errorf("validate id_token: %w", err)
	}

	// nonce 照合。これが無いと、他所で発行された正規の id_token を投げ込まれても気づけない。
	got := claimString(payload.Claims, "nonce")
	if got == nil || subtle.ConstantTimeCompare([]byte(*got), []byte(nonce)) != 1 {
		return domain.OAuthProfile{}, errors.New("id_token nonce mismatch")
	}

	profile := domain.OAuthProfile{
		Provider:          domain.ProviderGoogle,
		ProviderAccountID: payload.Subject, // sub
		Email:             claimString(payload.Claims, "email"),
		EmailVerified:     claimBool(payload.Claims, "email_verified"),
		AvatarURL:         claimString(payload.Claims, "picture"),
	}
	if name := claimString(payload.Claims, "name"); name != nil {
		profile.DisplayName = *name
	}

	if profile.ProviderAccountID == "" {
		return domain.OAuthProfile{}, errors.New("id_token has no sub claim")
	}

	return profile, nil
}

// idtoken.Payload の Claims は map[string]any なので、取り出しに型アサーションが要る。
// 内側の層へ持ち込まないよう、ここで型の付いた値に変える。
func claimString(claims map[string]any, key string) *string {
	s, ok := claims[key].(string)
	if !ok || s == "" {
		return nil
	}

	return &s
}

func claimBool(claims map[string]any, key string) bool {
	b, _ := claims[key].(bool)

	return b
}

// ----- 認証まわりの永続化 -----
// ここだけが GORM と Postgres の都合を知っている。

// AuthRepo は usecase.AuthRepository の実装。GORM 経由で Postgres を読み書きする。
// FindUserByOAuth は oauth_accounts から (provider, sub) でユーザーを引く。退会済みは「いない」扱い。
// CreateUserWithOAuth は users と oauth_accounts を 1 トランザクションで作る。
// CreateSession は sessions に 1 行入れる。
// FindUserBySessionToken は有効なセッションのユーザーを返し、最終アクセス時刻を更新する。
// RevokeSession は revoked_at を埋めて失効させる。
type AuthRepo struct {
	db *gorm.DB
}

var _ usecase.AuthRepository = (*AuthRepo)(nil)

func NewAuthRepo(db *gorm.DB) *AuthRepo {
	return &AuthRepo{db: db}
}

// FindUserByOAuth は (provider, sub) からユーザーを引く。
func (r *AuthRepo) FindUserByOAuth(ctx context.Context, provider, providerAccountID string) (*domain.User, error) {
	var account OAuthAccount

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("provider = ? AND provider_account_id = ?", provider, providerAccountID).
		First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select oauth_account: %w", err)
	}

	// 退会済み（匿名化済み）は存在しない扱いにする。
	if account.User == nil || account.User.AnonymizedAt != nil {
		return nil, domain.ErrUserNotFound
	}

	return toDomainUser(account.User), nil
}

// CreateUserWithOAuth はユーザーと OAuth 連携を 1 トランザクションで作る。
//
// 片方だけ残ると「ログインできないのにメールだけ埋まっている」幽霊ユーザーになるので、
// 必ず両方まとめて成功／失敗させる。
func (r *AuthRepo) CreateUserWithOAuth(ctx context.Context, profile domain.OAuthProfile, handle string) (*domain.User, error) {
	email := profile.TrustedEmail()

	user := User{
		ID:          uuid.New(),
		Email:       email,
		DisplayName: profile.SafeDisplayName(),
		Handle:      handle,
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return translateUniqueViolation(err)
		}

		account := OAuthAccount{
			ID:                uuid.New(),
			UserID:            user.ID,
			Provider:          profile.Provider,
			ProviderAccountID: profile.ProviderAccountID,
			ProviderEmail:     email,
		}
		if err := tx.Create(&account).Error; err != nil {
			return translateUniqueViolation(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return toDomainUser(&user), nil
}

func (r *AuthRepo) CreateSession(ctx context.Context, session domain.Session) error {
	row := Session{
		ID:         session.ID,
		UserID:     session.UserID,
		TokenHash:  session.TokenHash,
		ExpiresAt:  session.ExpiresAt,
		LastSeenAt: session.LastSeenAt,
	}

	if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

// FindUserBySessionToken は有効なセッションのユーザーを返し、最終アクセス時刻を更新する。
func (r *AuthRepo) FindUserBySessionToken(ctx context.Context, tokenHash string) (*domain.User, error) {
	var session Session

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > now()", tokenHash).
		First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select session: %w", err)
	}

	if session.User == nil || session.User.AnonymizedAt != nil || session.User.Status != userStatusActive {
		return nil, domain.ErrSessionNotFound
	}

	r.touchSession(ctx, session.ID)

	return toDomainUser(session.User), nil
}

func (r *AuthRepo) RevokeSession(ctx context.Context, tokenHash string) error {
	err := r.db.WithContext(ctx).
		Model(&Session{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", time.Now()).Error
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	// 対象 0 件でもエラーにしない。ログアウトは何度呼ばれても同じ結果であるべき。
	return nil
}

// touchSession は最終アクセス時刻だけを更新する。
// 失敗しても認証そのものは成立しているので、エラーは返さない。
func (r *AuthRepo) touchSession(ctx context.Context, id uuid.UUID) {
	r.db.WithContext(ctx).
		Model(&Session{}).
		Where("id = ?", id).
		Update("last_seen_at", time.Now())
}

const userStatusActive = "active"

// 一意制約の名前。migration の CREATE UNIQUE INDEX と揃える。
const (
	uxUsersHandle          = "ux_users_handle"
	uxUsersEmail           = "ux_users_email"
	uxOAuthProviderAccount = "ux_oauth_provider_account"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgUniqueViolation = "23505"

// translateUniqueViolation は Postgres 固有のエラーを domain のエラーに翻訳する。
//
// この変換をここでやるから、usecase 層は pgconn を import せずに
// errors.Is(err, domain.ErrHandleTaken) で分岐できる。
func translateUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgUniqueViolation {
		return fmt.Errorf("insert: %w", err)
	}

	switch pgErr.ConstraintName {
	case uxUsersHandle:
		return domain.ErrHandleTaken
	case uxUsersEmail:
		return domain.ErrEmailTaken
	case uxOAuthProviderAccount:
		return domain.ErrOAuthAccountTaken
	default:
		return fmt.Errorf("unique violation on %s: %w", pgErr.ConstraintName, err)
	}
}

// toDomainUser は DB 都合の型を、内側の層が扱う型に詰め替える。
func toDomainUser(u *User) *domain.User {
	return &domain.User{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Handle:      u.Handle,
		Status:      u.Status,
		Level:       u.Level,
		Rating:      u.Rating,
		CreatedAt:   u.CreatedAt,
	}
}
