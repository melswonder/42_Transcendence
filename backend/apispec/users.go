// users テーブル（§3）に対応する公開 API の仕様。
// 自分自身のプロフィール（/users/me）と、他人から見える公開プロフィールを型で分けている。
package apispec

import "time"

// UserPublic は他ユーザーから見えるプロフィール。
// email や preferred_locale のような個人設定は含めない。
type UserPublic struct {
	ID               string    `json:"id"                format:"uuid" example:"3f0c1a2e-9a5b-4c7d-8e1f-2b3c4d5e6f70"`
	DisplayName      string    `json:"display_name"      example:"melswonder"`
	Handle           string    `json:"handle"            example:"mels"`
	AvatarURL        string    `json:"avatar_url"        example:"https://cdn.example.com/avatars/default.png"` // 未設定ならデフォルト画像の URL
	Level            int       `json:"level"             example:"7"`
	ExperiencePoints int       `json:"experience_points" example:"1420"`
	Status           string    `json:"status"            enums:"active,suspended,deleted" example:"active"`
	CreatedAt        time.Time `json:"created_at"`
}

// UserPrivate は本人だけが読める自分のプロフィール。
// email は OAuth 専用ユーザーだと存在しないため null になりうる。
type UserPrivate struct {
	UserPublic
	Email           *string   `json:"email"            format:"email" example:"user@example.com"`
	PreferredLocale string    `json:"preferred_locale" example:"ja"`
	HasPassword     bool      `json:"has_password"     example:"true"`      // password_hash の有無だけを公開する
	LinkedProviders []string  `json:"linked_providers" example:"google,42"` // oauth_accounts から導出
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserListResponse は検索・一覧の共通レスポンス。
type UserListResponse struct {
	Items []UserPublic `json:"items"`
	Pagination
}

// UpdateMeRequest は自分のプロフィール更新。省略したフィールドは変更しない。
type UpdateMeRequest struct {
	DisplayName     *string `json:"display_name"     example:"melswonder"`
	Handle          *string `json:"handle"           example:"mels"`
	PreferredLocale *string `json:"preferred_locale" enums:"ja,en,fr" example:"ja"`
	AvatarAssetID   *string `json:"avatar_asset_id"  format:"uuid" example:"9b7f2c11-33aa-4d0e-9f21-77c1b0a5e123"` // null を渡すとアバターを外す
}

// DeleteMeRequest は退会（匿名化）の確認。パスワードを持つユーザーは本人確認のため必須。
type DeleteMeRequest struct {
	Password string `json:"password" example:"correct horse battery staple"`
}

// GetMe godoc
//
//	@Summary		自分のプロフィールを取得
//	@Description	アクセストークンの持ち主のプロフィールを返す。email や連携済み OAuth プロバイダを含む。
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UserPrivate
//	@Failure		401	{object}	ErrorResponse
//	@Router			/users/me [get]
func GetMe() {}

// UpdateMe godoc
//
//	@Summary		自分のプロフィールを更新
//	@Description	表示名・handle・言語・アバターを部分更新する。handle は退会者を除いて一意。
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		UpdateMeRequest	true	"更新する項目だけを含める"
//	@Success		200		{object}	UserPrivate
//	@Failure		400		{object}	ErrorResponse	"バリデーションエラー"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		409		{object}	ErrorResponse	"handle が既に使われている"
//	@Router			/users/me [patch]
func UpdateMe() {}

// DeleteMe godoc
//
//	@Summary		退会する（匿名化）
//	@Description	users.anonymized_at を立てて個人情報を落とす。handle と email は他ユーザーが再利用できるようになる。
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	DeleteMeRequest	false	"パスワードを持つユーザーは必須"
//	@Success		204		"退会完了。全セッションが失効する"
//	@Failure		401		{object}	ErrorResponse
//	@Router			/users/me [delete]
func DeleteMe() {}

// ListUsers godoc
//
//	@Summary		ユーザーを検索
//	@Description	handle / 表示名の部分一致で検索する。退会済みとブロック関係にあるユーザーは含めない。
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q		query		string	false	"handle または表示名の部分一致"	example(mels)
//	@Param			handle	query		string	false	"handle の完全一致"
//	@Param			limit	query		int		false	"取得件数 (1-100)"	default(20)	minimum(1)	maximum(100)
//	@Param			offset	query		int		false	"取得開始位置"		default(0)	minimum(0)
//	@Success		200		{object}	UserListResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/users [get]
func ListUsers() {}

// GetUser godoc
//
//	@Summary		ユーザーの公開プロフィールを取得
//	@Tags			users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			userId	path		string	true	"ユーザー ID"	format(uuid)
//	@Success		200		{object}	UserPublic
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse	"存在しない、または退会済み"
//	@Router			/users/{userId} [get]
func GetUser() {}
