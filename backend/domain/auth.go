package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

// パスワードの長さ制限。上限は bcrypt が 72 バイトまでしか見ないため。
const (
	PasswordMinLen = 8
	PasswordMaxLen = 72
)

// ValidateEmail は形式だけを軽く確かめる。厳密な検証はしない
// （本人確認はメール到達でしかできず、一意性は DB の制約が守る）。
func ValidateEmail(email string) error {
	at := strings.IndexByte(email, '@')
	if at < 1 || at == len(email)-1 || len(email) > 254 || strings.ContainsAny(email, " \t") {
		return ErrInvalidEmail
	}
	return nil
}

// ValidatePassword は長さだけを見る。文字種の強制はしない（長さが最も効く）。
func ValidatePassword(password string) error {
	if len(password) < PasswordMinLen || len(password) > PasswordMaxLen {
		return ErrWeakPassword
	}
	return nil
}

// HashPassword は保存用のハッシュを作る。bcrypt はソルトを内包する。
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword はハッシュと平文を照合する。
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// dummyPasswordHash は存在しないユーザーへのログイン試行でも
// 同じだけ時間を使うためのダミー。ユーザーの有無を応答時間で悟らせない。
var dummyPasswordHash, _ = HashPassword("dummy password for timing")

// CheckPasswordDummy はダミー照合。結果は常に false。
func CheckPasswordDummy() {
	_ = CheckPassword(dummyPasswordHash, "not the password")
}
