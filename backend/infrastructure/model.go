package infrastructure

import (
	"time"

	"github.com/google/uuid"
)

// User - users (§3)
//
// ux_users_email / ux_users_handle は WHERE 付きの部分ユニークインデックス。
// 退会（匿名化）した行を対象から外すことで、退会者が握っていた handle / email を再利用できる。
//
// @migration CHECK (password_hash IS NOT NULL OR email IS NOT NULL OR anonymized_at IS NOT NULL)
// @migration FK avatar_asset_id -> media_assets(id)  ※循環参照のため後付け
type User struct {
	ID               uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	Email            *string    `gorm:"column:email;type:citext;uniqueIndex:ux_users_email,where:email IS NOT NULL AND anonymized_at IS NULL"` // OAuthがメールを返さない場合があるためNULL許容
	PasswordHash     *string    `gorm:"column:password_hash;type:text"`                                                                        // OAuth専用ユーザーは持たない
	DisplayName      string     `gorm:"column:display_name;type:varchar(50);not null"`
	Handle           string     `gorm:"column:handle;type:varchar(30);not null;uniqueIndex:ux_users_handle,where:anonymized_at IS NULL"`
	AvatarAssetID    *uuid.UUID `gorm:"column:avatar_asset_id;type:uuid"` // FKは手書きmigrationで付与
	PreferredLocale  string     `gorm:"column:preferred_locale;type:varchar(10);not null;default:ja"`
	Status           string     `gorm:"column:status;type:varchar(20);not null;default:active;check:status IN ('active','suspended','deleted')"`
	Level            int        `gorm:"column:level;type:integer;not null;default:1;check:level >= 1"`
	ExperiencePoints int        `gorm:"column:experience_points;type:integer;not null;default:0;check:experience_points >= 0"`
	Rating           int        `gorm:"column:rating;type:integer;not null;default:1200;index:idx_users_rating,sort:desc"` // Elo。Leaderboard はこの列だけで並べる
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

// Match - matches
//
// 1 対戦ぶんの事実だけを持ち、誰が勝ったかは MatchParticipant 側で表す。
// 進行中は status='in_progress' / result_type と finished_at が NULL。
//
// @migration CHECK (status <> 'finished' OR (result_type IS NOT NULL AND finished_at IS NOT NULL))
type Match struct {
	ID         uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	Mode       string     `gorm:"column:mode;type:varchar(20);not null;check:mode IN ('ranked','casual','ai','friend')"`
	Status     string     `gorm:"column:status;type:varchar(20);not null;default:in_progress;index:idx_matches_status_finished,priority:1;check:status IN ('in_progress','finished','aborted')"`
	ResultType *string    `gorm:"column:result_type;type:varchar(20);check:result_type IS NULL OR result_type IN ('goal','resign','timeout','draw','abort')"` // 決着のつき方
	TotalMoves int        `gorm:"column:total_moves;type:integer;not null;default:0;check:total_moves >= 0"`
	StartedAt  time.Time  `gorm:"column:started_at;type:timestamptz;not null"`
	FinishedAt *time.Time `gorm:"column:finished_at;type:timestamptz;index:idx_matches_status_finished,priority:2,sort:desc"`
	CreatedAt  time.Time  `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
}

func (Match) TableName() string { return "matches" }

// MatchParticipant - match_participants
//
// 1 対戦につき 2 行。統計はすべてこのテーブルの user_id で引く。
// matches 側に player1_id / player2_id を持たせないのは、そうすると勝敗集計が
// 「自分が player1 の場合と player2 の場合」の 2 通りに分岐してしまうため。
//
// 対局開始時に outcome / rating_after を NULL のまま挿入し、決着時に埋める。
// 進行中の対局へ再接続するとき、この 2 行から座席と参加者を復元する。
//
// rating_before / rating_after はレーティング推移のグラフに使う。users.rating は
// 最新値しか持たないので、履歴はここに残す。
//
// @migration CHECK (rating_before >= 0 AND (rating_after IS NULL OR rating_after >= 0))
type MatchParticipant struct {
	MatchID      uuid.UUID `gorm:"column:match_id;type:uuid;primaryKey"`
	UserID       uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey;index:idx_match_participants_user,priority:1"`
	Seat         int16     `gorm:"column:seat;type:smallint;not null;check:seat IN (0,1)"`                                    // 0=先手 1=後手
	Outcome      *string   `gorm:"column:outcome;type:varchar(10);check:outcome IS NULL OR outcome IN ('win','loss','draw')"` // 進行中は NULL
	RatingBefore int       `gorm:"column:rating_before;type:integer;not null"`
	RatingAfter  *int      `gorm:"column:rating_after;type:integer"` // 進行中は NULL
	XPGained     int       `gorm:"column:xp_gained;type:integer;not null;default:0;check:xp_gained >= 0"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`

	Match *Match `gorm:"foreignKey:MatchID;references:ID"`
	User  *User  `gorm:"foreignKey:UserID;references:ID"`
}

func (MatchParticipant) TableName() string { return "match_participants" }

// MatchAction - match_actions
//
// 進行中の対局の 1 手（駒移動・壁配置・投了など）を順番に積む追記専用ログ。
// 盤面そのものは保存せず、NewQuoridor() にこのログを順に適用して復元する。
//
// action_seq は 1 から始まる連番で、そのままゲームの版数（gameVersion）になる。
// クライアントは「いま自分が見ている版数」を添えて操作を送り、
// 合わなければ古い操作として拒否する（楽観ロック）。
//
// action_id はクライアントが生成する冪等キー。再送で同じ操作が 2 回届いても
// (match_id, action_id) のユニーク制約で 2 回目を検出し、適用済みとして扱う。
type MatchAction struct {
	MatchID    uuid.UUID `gorm:"column:match_id;type:uuid;primaryKey;uniqueIndex:ux_match_actions_action_id,priority:1"`
	ActionSeq  int       `gorm:"column:action_seq;type:integer;primaryKey;check:action_seq >= 1"`
	ActionID   uuid.UUID `gorm:"column:action_id;type:uuid;not null;uniqueIndex:ux_match_actions_action_id,priority:2"`
	ActorSeat  int16     `gorm:"column:actor_seat;type:smallint;not null;check:actor_seat IN (0,1)"`
	ActionType string    `gorm:"column:action_type;type:varchar(20);not null;check:action_type IN ('move','wall','resign','timeout','abort')"`
	Payload    string    `gorm:"column:payload;type:jsonb;not null;default:'{}'"` // 座標などの中身。形は domain 側の型で決める
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`

	Match *Match `gorm:"foreignKey:MatchID;references:ID"`
}

func (MatchAction) TableName() string { return "match_actions" }

// UserAchievement - user_achievements
//
// 解除済みの実績だけを持つ。実績の定義（名前・説明・達成条件）は
// domain/achievement.go 側に定数で置く。マスタを DB に入れると実績を 1 つ足すたびに
// データ投入の migration が要るため。
type UserAchievement struct {
	UserID     uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey"` // PK の先頭列なので単独の索引は張らない
	Code       string    `gorm:"column:code;type:varchar(50);primaryKey"`
	UnlockedAt time.Time `gorm:"column:unlocked_at;type:timestamptz;not null;autoCreateTime"`

	User *User `gorm:"foreignKey:UserID;references:ID"`
}

func (UserAchievement) TableName() string { return "user_achievements" }
