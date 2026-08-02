// GORM の永続化モデルを置く。docs/database-design.md を正本とし、それをGoの構造体へ写したもの。
//
// domain 層ではなく infrastructure 層に置くのは、GORMタグがDBの都合であり、
// domain は「何にも依存しない」層だから（Issue #6 の受け入れ条件）。
// domain のエンティティとは別物として扱い、変換は repository の責務とする。
//
// 設計書の制約のうち、以下はGORMのタグでは表現できないため手書きSQL migrationで補う
// （docs/database-design.md §13）。
//
//   - partial unique index（WHERE句付きのUNIQUE）
//   - CREATE EXTENSION citext
//   - users ⇄ media_assets の循環FK（ALTER TABLE で後付けが必要）
//
// したがって cmd/migrate が出力するDDLは土台であって完成形ではない。
package infrastructure

import (
	"time"

	"github.com/google/uuid"
)

// User は docs/database-design.md §3 users に対応する。
//
// 手書きmigrationで追加が必要なもの:
//   - CREATE UNIQUE INDEX ux_users_email  ON users(email)  WHERE email IS NOT NULL AND anonymized_at IS NULL;
//   - CREATE UNIQUE INDEX ux_users_handle ON users(handle) WHERE anonymized_at IS NULL;
//   - CHECK (password_hash IS NOT NULL OR email IS NOT NULL OR anonymized_at IS NOT NULL)
//   - avatar_asset_id の FK（MediaAsset との循環参照のため）
type User struct {
	ID uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	// OAuthプロバイダがメールを返さない場合があるためNULL許容。
	Email *string `gorm:"column:email;type:citext"`
	// OAuth専用ユーザーはパスワードを持たないためNULL許容。
	PasswordHash *string `gorm:"column:password_hash;type:text"`
	DisplayName  string  `gorm:"column:display_name;type:varchar(50);not null"`
	Handle       string  `gorm:"column:handle;type:varchar(30);not null"`
	// FK制約は循環参照のため手書きmigrationで後付けする。
	AvatarAssetID    *uuid.UUID `gorm:"column:avatar_asset_id;type:uuid"`
	PreferredLocale  string     `gorm:"column:preferred_locale;type:varchar(10);not null;default:ja"`
	Status           string     `gorm:"column:status;type:varchar(20);not null;default:active;check:status IN ('active','suspended','deleted')"`
	Level            int        `gorm:"column:level;type:integer;not null;default:1;check:level >= 1"`
	ExperiencePoints int        `gorm:"column:experience_points;type:integer;not null;default:0;check:experience_points >= 0"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
	AnonymizedAt     *time.Time `gorm:"column:anonymized_at;type:timestamptz"`
}

func (User) TableName() string { return "users" }

// MediaAsset は docs/database-design.md §3 media_assets に対応する。
// アバター画像のアップロード管理に使う。未設定時はレコードを作らず、
// frontend/backend共通のデフォルトアバターURLを返す。
type MediaAsset struct {
	ID          uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	OwnerUserID uuid.UUID `gorm:"column:owner_user_id;type:uuid;not null;index:idx_media_assets_owner,priority:1"`
	Purpose     string    `gorm:"column:purpose;type:varchar(30);not null;index:idx_media_assets_owner,priority:2;check:purpose IN ('avatar')"`
	// 推測困難な値を入れる。URLから他人のアセットを列挙できないようにするため。
	StorageKey       string     `gorm:"column:storage_key;type:text;not null;uniqueIndex"`
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

// OAuthAccount は docs/database-design.md §3 oauth_accounts に対応する。
type OAuthAccount struct {
	ID     uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID uuid.UUID `gorm:"column:user_id;type:uuid;not null;index"`
	// provider identity の重複を防ぐ。同じGitHubアカウントで2ユーザーは作れない。
	Provider          string    `gorm:"column:provider;type:varchar(30);not null;uniqueIndex:ux_oauth_provider_account,priority:1;check:provider IN ('google','github','42')"`
	ProviderAccountID string    `gorm:"column:provider_account_id;type:varchar(255);not null;uniqueIndex:ux_oauth_provider_account,priority:2"`
	ProviderEmail     *string   `gorm:"column:provider_email;type:citext"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

func (OAuthAccount) TableName() string { return "oauth_accounts" }

// Session は docs/database-design.md §3 sessions に対応する。
// raw token は保存せず、ハッシュのみを持つ。
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

// Friendship は docs/database-design.md §3 friendships に対応する。
//
// (A,B) と (B,A) の二重登録を防ぐため user_low_id < user_high_id に正規化する。
// アプリ側は常に LEAST/GREATEST で並べ替えてから読み書きすること。
// CHECK (user_low_id < user_high_id) は手書きmigrationで追加する。
type Friendship struct {
	UserLowID  uuid.UUID `gorm:"column:user_low_id;type:uuid;primaryKey"`
	UserHighID uuid.UUID `gorm:"column:user_high_id;type:uuid;primaryKey;index:idx_friendships_high_status,priority:1"`
	// どちらが申請したかを保持する。正規化により行の並びからは判別できないため。
	RequestedByUserID uuid.UUID `gorm:"column:requested_by_user_id;type:uuid;not null"`
	Status            string    `gorm:"column:status;type:varchar(20);not null;index:idx_friendships_high_status,priority:2;check:status IN ('pending','accepted','rejected')"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`

	UserLow  *User `gorm:"foreignKey:UserLowID;references:ID"`
	UserHigh *User `gorm:"foreignKey:UserHighID;references:ID"`
}

func (Friendship) TableName() string { return "friendships" }

// Block は docs/database-design.md §3 blocks に対応する。
//
// ブロックは方向を持つ関係なので Friendship とは別テーブルにする。
// 「AがBをブロックしたが、BはAをブロックしていない」を表現するため。
type Block struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	BlockerUserID uuid.UUID `gorm:"column:blocker_user_id;type:uuid;not null;uniqueIndex:ux_blocks_pair,priority:1"`
	BlockedUserID uuid.UUID `gorm:"column:blocked_user_id;type:uuid;not null;uniqueIndex:ux_blocks_pair,priority:2;index"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`

	Blocker *User `gorm:"foreignKey:BlockerUserID;references:ID"`
	Blocked *User `gorm:"foreignKey:BlockedUserID;references:ID"`
}

func (Block) TableName() string { return "blocks" }
