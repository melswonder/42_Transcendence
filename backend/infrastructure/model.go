// Package infrastructure holds the GORM persistence models.
//
// 正本は docs/database-design.md。GORMタグで表現できないもの
// （partial unique index / CREATE EXTENSION citext / users ⇄ media_assets の循環FK）は
// 手書きSQL migrationで補う（§13）。以下 @migration 印がそれ。
package infrastructure

import (
	"time"

	"github.com/google/uuid"
)

// User - users (§3)
//
// @migration ux_users_email  UNIQUE(email)  WHERE email IS NOT NULL AND anonymized_at IS NULL
// @migration ux_users_handle UNIQUE(handle) WHERE anonymized_at IS NULL
// @migration CHECK (password_hash IS NOT NULL OR email IS NOT NULL OR anonymized_at IS NOT NULL)
// @migration FK avatar_asset_id -> media_assets(id)  ※循環参照のため後付け
type User struct {
	ID               uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	Email            *string    `gorm:"column:email;type:citext"`       // OAuthがメールを返さない場合があるためNULL許容
	PasswordHash     *string    `gorm:"column:password_hash;type:text"` // OAuth専用ユーザーは持たない
	DisplayName      string     `gorm:"column:display_name;type:varchar(50);not null"`
	Handle           string     `gorm:"column:handle;type:varchar(30);not null"`
	AvatarAssetID    *uuid.UUID `gorm:"column:avatar_asset_id;type:uuid"` // FKは手書きmigrationで付与
	PreferredLocale  string     `gorm:"column:preferred_locale;type:varchar(10);not null;default:ja"`
	Status           string     `gorm:"column:status;type:varchar(20);not null;default:active;check:status IN ('active','suspended','deleted')"`
	Level            int        `gorm:"column:level;type:integer;not null;default:1;check:level >= 1"`
	ExperiencePoints int        `gorm:"column:experience_points;type:integer;not null;default:0;check:experience_points >= 0"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
	AnonymizedAt     *time.Time `gorm:"column:anonymized_at;type:timestamptz"`
}

func (User) TableName() string { return "users" }

// MediaAsset - media_assets (§3)
// アバター画像のアップロード管理。未設定時はレコードを作らずデフォルトアバターURLを返す。
type MediaAsset struct {
	ID               uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerUserID      uuid.UUID  `gorm:"column:owner_user_id;type:uuid;not null;index:idx_media_assets_owner,priority:1"`
	Purpose          string     `gorm:"column:purpose;type:varchar(30);not null;index:idx_media_assets_owner,priority:2;check:purpose IN ('avatar')"`
	StorageKey       string     `gorm:"column:storage_key;type:text;not null;uniqueIndex"` // 推測困難な値。URLからの列挙を防ぐ
	OriginalFilename string     `gorm:"column:original_filename;type:varchar(255);not null"`
	MimeType         string     `gorm:"column:mime_type;type:varchar(100);not null"`
	SizeBytes        int64      `gorm:"column:size_bytes;type:bigint;not null;check:size_bytes > 0"`
	Width            *int       `gorm:"column:width;type:integer"`
	Height           *int       `gorm:"column:height;type:integer"`
	ChecksumSHA256   string     `gorm:"column:checksum_sha256;type:char(64);not null"`
	Status           string     `gorm:"column:status;type:varchar(20);not null;default:active;index:idx_media_assets_owner,priority:3;check:status IN ('active','deleted')"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;type:timestamptz"`

	Owner *User `gorm:"foreignKey:OwnerUserID;references:ID"`
}

func (MediaAsset) TableName() string { return "media_assets" }

// OAuthAccount - oauth_accounts (§3)
type OAuthAccount struct {
	ID                uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID            uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"`
	Provider          string    `gorm:"column:provider;type:varchar(30);not null;uniqueIndex:ux_oauth_provider_account,priority:1;check:provider IN ('google','github','42')"` // (provider, account_id) で一意
	ProviderAccountID string    `gorm:"column:provider_account_id;type:varchar(255);not null;uniqueIndex:ux_oauth_provider_account,priority:2"`
	ProviderEmail     *string   `gorm:"column:provider_email;type:citext"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

func (OAuthAccount) TableName() string { return "oauth_accounts" }

// Session - sessions (§3)
// raw token は保存せずハッシュのみ持つ。
type Session struct {
	ID         uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID  `gorm:"column:user_id;type:uuid;not null;index:idx_sessions_user_expires,priority:1"`
	TokenHash  string     `gorm:"column:token_hash;type:text;not null;uniqueIndex"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;type:timestamptz;not null;index:idx_sessions_user_expires,priority:2"`
	LastSeenAt time.Time  `gorm:"column:last_seen_at;type:timestamptz;not null"`
	RevokedAt  *time.Time `gorm:"column:revoked_at;type:timestamptz"`
	CreatedAt  time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

func (Session) TableName() string { return "sessions" }

// Friendship - friendships (§3)
//
// (A,B)/(B,A) の二重登録を防ぐため user_low_id < user_high_id に正規化する。
// 読み書きの前に必ず LEAST/GREATEST で並べ替えること。
//
// @migration CHECK (user_low_id < user_high_id)
type Friendship struct {
	UserLowID         uuid.UUID `gorm:"column:user_low_id;type:uuid;primaryKey"`
	UserHighID        uuid.UUID `gorm:"column:user_high_id;type:uuid;primaryKey;index:idx_friendships_high_status,priority:1"`
	RequestedByUserID uuid.UUID `gorm:"column:requested_by_user_id;type:uuid;not null"` // 正規化後は行の並びから申請者を判別できないため保持
	Status            string    `gorm:"column:status;type:varchar(20);not null;index:idx_friendships_high_status,priority:2;check:status IN ('pending','accepted','rejected')"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`

	UserLow  *User `gorm:"foreignKey:UserLowID;references:ID"`
	UserHigh *User `gorm:"foreignKey:UserHighID;references:ID"`
}

func (Friendship) TableName() string { return "friendships" }

// Block - blocks (§3)
// 方向を持つ関係のため Friendship とは別テーブル（片側だけのブロックを表現する）。
type Block struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	BlockerUserID uuid.UUID `gorm:"column:blocker_user_id;type:uuid;not null;uniqueIndex:ux_blocks_pair,priority:1"`
	BlockedUserID uuid.UUID `gorm:"column:blocked_user_id;type:uuid;not null;uniqueIndex:ux_blocks_pair,priority:2;index"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`

	Blocker *User `gorm:"foreignKey:BlockerUserID;references:ID"`
	Blocked *User `gorm:"foreignKey:BlockedUserID;references:ID"`
}

func (Block) TableName() string { return "blocks" }
