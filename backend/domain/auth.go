package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// 対応する外部プロバイダ。oauth_accounts.provider の CHECK 制約と同じ値にする。
const (
	ProviderGoogle = "google"
	ProviderGitHub = "github"
	Provider42     = "42"
)

// OAuthProfile は外部プロバイダから受け取った本人確認の結果。ライブラリの型を内側へ持ち込まないための、こちら側の言葉。
// SafeDisplayName は表示名を users.display_name に収まる形へ整える。空なら既定値。
// TrustedEmail は「確認済み」と言えるメールだけを返す。未確認なら nil。
type OAuthProfile struct {
	Provider          string  // "google"
	ProviderAccountID string  // Google の sub
	Email             *string // 来ないことがあるのでポインタ
	EmailVerified     bool
	DisplayName       string
	AvatarURL         *string
}

// Session はログイン状態そのもの。
//
// 生トークンは Cookie にだけ置き、DB にはハッシュを保存する。
// DB が漏れても、その中身だけでは誰にもなりすませないようにするため。
type Session struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	TokenHash  string
	ExpiresAt  time.Time
	LastSeenAt time.Time
}

// SessionTTL はログインが保たれる期間。
const SessionTTL = 7 * 24 * time.Hour

// NewSessionToken は Cookie に載せる生トークンを作る。
// rand.Text は crypto/rand 由来の base32 文字列（約 128bit）。
func NewSessionToken() string {
	return rand.Text()
}

// HashSessionToken は DB に保存する形へ変換する。
//
// パスワードと違いトークンは十分な長さのランダム値なので、
// bcrypt のような意図的に遅いハッシュは要らない。毎リクエスト叩くぶん速さが要る。
func HashSessionToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))

	return hex.EncodeToString(sum[:])
}

// NewSession は生トークンから保存用のセッションを組み立てる。
func NewSession(userID uuid.UUID, rawToken string, now time.Time) Session {
	return Session{
		ID:         uuid.New(),
		UserID:     userID,
		TokenHash:  HashSessionToken(rawToken),
		ExpiresAt:  now.Add(SessionTTL),
		LastSeenAt: now,
	}
}

// NewOAuthState は CSRF 対策の state と、id_token 紐付け用の nonce を作る。
//
//	state: 「この callback は自分が始めたログインの続きか」を確かめる
//	nonce: 「この id_token は自分が今出したリクエストへの返答か」を確かめる
func NewOAuthState() (state, nonce string) {
	return rand.Text(), rand.Text()
}
