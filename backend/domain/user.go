package domain

import (
	"crypto/rand"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User はアプリケーションが扱うユーザーそのもの。
// カラムの型や GORM のタグは infrastructure.User が持つ。この層は DB を知らない。
type User struct {
	ID               uuid.UUID
	Email            *string
	DisplayName      string
	Handle           string
	Status           string
	Level            int
	ExperiencePoints int
	Rating           int
	PreferredLocale  string
	AvatarAssetID    *uuid.UUID // 未設定ならデフォルトアバター
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Profile は自分のプロフィール表示に足す、User 本体の外にある情報。
type Profile struct {
	User
	HasPassword     bool     // password_hash の有無だけ公開する
	LinkedProviders []string // 連携済みの OAuth プロバイダ
}

// 対応するロケール。users.preferred_locale に入れる値。
var SupportedLocales = []string{"ja", "en", "fr"}

// ProfileUpdate はプロフィール編集の入力。nil のフィールドは変更しない。
// AvatarAssetID だけは「null で外す」があるので AvatarSet で有無を区別する。
type ProfileUpdate struct {
	DisplayName     *string
	Handle          *string
	PreferredLocale *string
	AvatarSet       bool
	AvatarAssetID   *uuid.UUID // AvatarSet かつ nil ならアバターを外す
}

// Validate は形式だけを確かめる。handle の一意性は DB の制約に任せる。
func (u ProfileUpdate) Validate() error {
	if u.DisplayName != nil {
		name := strings.TrimSpace(*u.DisplayName)
		if name == "" || len([]rune(name)) > DisplayNameMaxLen {
			return ErrInvalidDisplayName
		}
	}
	if u.Handle != nil {
		if err := ValidateHandle(*u.Handle); err != nil {
			return err
		}
	}
	if u.PreferredLocale != nil && !slices.Contains(SupportedLocales, *u.PreferredLocale) {
		return ErrInvalidLocale
	}
	return nil
}

// ValidateHandle は handle として使える形かを確かめる。
func ValidateHandle(handle string) error {
	if len(handle) < handleMinLen || len(handle) > HandleMaxLen {
		return ErrInvalidHandle
	}
	for _, r := range handle {
		if !isHandleRune(r) {
			return ErrInvalidHandle
		}
	}
	return nil
}

// DB のカラム長に合わせた上限 (users.display_name / users.handle)。
const (
	DisplayNameMaxLen = 50
	HandleMaxLen      = 30
	handleMinLen      = 3

	// 生成した handle が既存と衝突したときに作り直す回数。
	HandleGenerateAttempts = 8

	defaultDisplayName = "Player"
	defaultHandleBase  = "player"
)

// handle は URL (/users/xxx) とメンション (@xxx) に出るので、
// 英小文字・数字・アンダースコアだけに制限する。
func isHandleRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'
}

// HandleBase は OAuth プロフィールから handle の元ネタを作る。
//
// Google は handle に相当する値を返さないので、こちらで組み立てるしかない。
// メールのローカル部 → 表示名 → 固定値、の順に使えるものを探す。
func HandleBase(p OAuthProfile) string {
	candidates := []string{}
	if p.Email != nil {
		local, _, _ := strings.Cut(*p.Email, "@")
		candidates = append(candidates, local)
	}
	candidates = append(candidates, p.DisplayName, defaultHandleBase)

	for _, c := range candidates {
		if base := sanitizeHandle(c); len(base) >= handleMinLen {
			return base
		}
	}

	return defaultHandleBase
}

// sanitizeHandle は使えない文字を落とし、長さを詰める。
func sanitizeHandle(raw string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case isHandleRune(r):
			b.WriteRune(r)
		case r == '.' || r == '-' || r == ' ':
			// 区切り文字はアンダースコアに寄せる
			b.WriteRune('_')
		}
	}

	s := strings.Trim(b.String(), "_")
	if len(s) > HandleMaxLen {
		s = strings.TrimRight(s[:HandleMaxLen], "_")
	}

	return s
}

// HandleCandidate は attempt 回目の候補を返す。
//
// 1 回目は base をそのまま使い、衝突したら以降はランダムな接尾辞を付ける。
// 連番 (_2, _3...) にしないのは、他人の handle を推測しにくくするため。
func HandleCandidate(base string, attempt int) string {
	if attempt <= 0 {
		return base
	}

	const suffixLen = 4
	// "_" + 接尾辞 の分だけ base を削ってから足す。
	trimTo := HandleMaxLen - suffixLen - 1
	if len(base) > trimTo {
		base = strings.TrimRight(base[:trimTo], "_")
	}

	// rand.Text は base32 の文字列を返す。小文字化して handle の文字種に揃える。
	return base + "_" + strings.ToLower(rand.Text()[:suffixLen])
}

// SafeDisplayName は表示名を users.display_name に収まる形へ整える。
// Google が name を返さないケース（scope に profile を入れ忘れた等）にも備える。
func (p OAuthProfile) SafeDisplayName() string {
	name := strings.TrimSpace(p.DisplayName)
	if name == "" {
		return defaultDisplayName
	}

	// varchar(50) は文字数で数えるのでルーン単位で切る。
	if r := []rune(name); len(r) > DisplayNameMaxLen {
		return string(r[:DisplayNameMaxLen])
	}

	return name
}

// TrustedEmail は「確認済み」と言えるメールだけを返す。
//
// email_verified が false のメールを信じると、
// 攻撃者が未確認の他人のメールでアカウントを作れてしまう。
func (p OAuthProfile) TrustedEmail() *string {
	if p.Email == nil || !p.EmailVerified {
		return nil
	}

	return p.Email
}
