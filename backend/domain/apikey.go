package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Public API のスコープ。キー作成時に選び、エンドポイントごとに要求される。
const (
	APIScopeRead  = "read"  // 参照系（プロフィール・履歴・統計・ランキング・フレンド一覧）
	APIScopeWrite = "write" // 更新系（プロフィール編集・フレンド申請/解除）
)

var APIScopes = []string{APIScopeRead, APIScopeWrite}

// APIKeyNameMaxLen は api_keys.name の上限。
const APIKeyNameMaxLen = 50

// apiKeyPrefix は raw key の先頭に付く目印。ログや issue に紛れても見分けられる。
const apiKeyPrefix = "tsc_"

// Public API まわりの失敗。handler は errors.Is で 401 / 403 / 429 に振り分ける。
var (
	ErrAPIKeyNotFound     = errors.New("api key not found")
	ErrAPIKeyInvalid      = errors.New("invalid api key")
	ErrAPIKeyExpired      = errors.New("api key expired")
	ErrAPIKeyRevoked      = errors.New("api key revoked")
	ErrAPIKeyScope        = errors.New("api key lacks required scope")
	ErrAPIKeyRateLimited  = errors.New("rate limit exceeded")
	ErrInvalidAPIKeyInput = errors.New("invalid api key input")
)

// APIKey は発行済みのキー 1 本。raw key は持たない（ハッシュだけを保存する）。
type APIKey struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Name       string
	KeyPrefix  string // raw key の先頭数文字。一覧でどのキーか見分ける用
	Scopes     []string
	ExpiresAt  *time.Time // nil なら無期限
	RevokedAt  *time.Time
	LastUsedAt *time.Time
	CreatedAt  time.Time
}

func (k APIKey) Revoked() bool { return k.RevokedAt != nil }

func (k APIKey) Expired(now time.Time) bool {
	return k.ExpiresAt != nil && now.After(*k.ExpiresAt)
}

func (k APIKey) HasScope(scope string) bool {
	return slices.Contains(k.Scopes, scope)
}

// NewAPIKeySecret は raw key を作る。crypto/rand 由来 ~208bit。
// DB にはハッシュしか置かないので、この値は作成レスポンスで一度だけ見せる。
func NewAPIKeySecret() string {
	return apiKeyPrefix + strings.ToLower(rand.Text()+rand.Text())
}

// HashAPIKey は保存・照合用のハッシュ。セッショントークンと同じ考え方で、
// 十分に長いランダム値なので遅いハッシュは要らない。
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// APIKeyPrefixOf は一覧表示用に raw key の先頭を切り出す。
func APIKeyPrefixOf(raw string) string {
	const visible = len(apiKeyPrefix) + 6
	if len(raw) < visible {
		return raw
	}
	return raw[:visible]
}

func ValidateAPIKeyInput(name string, scopes []string, expiresAt *time.Time, now time.Time) error {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > APIKeyNameMaxLen {
		return ErrInvalidAPIKeyInput
	}
	if len(scopes) == 0 {
		return ErrInvalidAPIKeyInput
	}
	for _, s := range scopes {
		if !slices.Contains(APIScopes, s) {
			return ErrInvalidAPIKeyInput
		}
	}
	if expiresAt != nil && !expiresAt.After(now) {
		return ErrInvalidAPIKeyInput
	}
	return nil
}
