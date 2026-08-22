// Public API（/v1）と API キー管理の仕様。
//
// /v1 配下は Cookie ではなく Authorization: Bearer <api key> で認証する。
// キーには read / write のスコープと任意の有効期限があり、
// 失効・期限切れ・スコープ不足はそれぞれ別のコードで拒否される。
// 流量はキー単位（既定 60 リクエスト/分）で、状態は X-RateLimit-* ヘッダで返す。
package apispec

import "time"

// APIKeyResponse は発行済みキーの 1 本。raw key は含まれない。
type APIKeyResponse struct {
	ID         string     `json:"id"           format:"uuid"`
	Name       string     `json:"name"         example:"my analytics bot"`
	KeyPrefix  string     `json:"key_prefix"   example:"tsc_a1b2c3"` // どのキーか見分ける用。認証には使えない
	Scopes     []string   `json:"scopes"       example:"read,write"` // read / write
	ExpiresAt  *time.Time `json:"expires_at"`                        // null なら無期限
	RevokedAt  *time.Time `json:"revoked_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// APIKeyCreatedResponse は作成時だけ raw key を含む。DB はハッシュ保存なので二度と取得できない。
type APIKeyCreatedResponse struct {
	APIKey string         `json:"api_key" example:"tsc_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`
	Key    APIKeyResponse `json:"key"`
}

// APIKeyCreateRequest はキー作成の入力。
type APIKeyCreateRequest struct {
	Name      string     `json:"name"       binding:"required" example:"my analytics bot"`
	Scopes    []string   `json:"scopes"     binding:"required" example:"read"` // read / write
	ExpiresAt *time.Time `json:"expires_at"`                                   // 省略で無期限
}

// APIKeyListResponse は自分のキー一覧。
type APIKeyListResponse struct {
	Items []APIKeyResponse `json:"items"`
}

// PublicProfileUpdateRequest は Public API からのプロフィール編集。省略したフィールドは変更しない。
type PublicProfileUpdateRequest struct {
	DisplayName     *string `json:"display_name"     example:"Alice"`
	Handle          *string `json:"handle"           example:"alice_42"`
	PreferredLocale *string `json:"preferred_locale" enums:"ja,en,fr"`
}

// CreateAPIKey は API キーを発行する。
//
//	@Summary	API キーを発行する
//	@Description	raw key はこのレスポンスで一度だけ返す。DB にはハッシュしか保存しない。
//	@Tags		apikeys
//	@Accept		json
//	@Produce	json
//	@Param		request	body		APIKeyCreateRequest	true	"作成内容"
//	@Success	201		{object}	APIKeyCreatedResponse
//	@Failure	400		{object}	ErrorResponse	"名前・スコープ・期限が不正"
//	@Failure	401		{object}	ErrorResponse	"未ログイン（管理は Cookie セッション）"
//	@Router		/apikeys [post]
func CreateAPIKey() {}

// ListAPIKeys は自分のキー一覧を返す。
//
//	@Summary	API キー一覧
//	@Tags		apikeys
//	@Produce	json
//	@Success	200	{object}	APIKeyListResponse
//	@Failure	401	{object}	ErrorResponse
//	@Router		/apikeys [get]
func ListAPIKeys() {}

// RevokeAPIKey はキーを失効させる。
//
//	@Summary	API キーを失効させる
//	@Description	行は消さず revoked_at を立てる。以後そのキーのリクエストは 401 になる。
//	@Tags		apikeys
//	@Param		keyId	path	string	true	"キー ID"	format(uuid)
//	@Success	204
//	@Failure	401	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse	"自分のキーに存在しない"
//	@Router		/apikeys/{keyId} [delete]
func RevokeAPIKey() {}

// PublicGetProfile はキー所有者のプロフィールを返す。
//
//	@Summary	自分のプロフィール（Public API）
//	@Tags		public
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	UserPrivate
//	@Failure	401	{object}	ErrorResponse	"invalid_api_key / api_key_expired / api_key_revoked"
//	@Failure	403	{object}	ErrorResponse	"insufficient_scope（read が必要）"
//	@Failure	429	{object}	ErrorResponse	"rate_limited"
//	@Router		/v1/profile [get]
func PublicGetProfile() {}

// PublicUpdateProfile はプロフィールを編集する。
//
//	@Summary	プロフィール編集（Public API）
//	@Tags		public
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		PublicProfileUpdateRequest	true	"変更内容"
//	@Success	200		{object}	UserPrivate
//	@Failure	400		{object}	ErrorResponse
//	@Failure	401		{object}	ErrorResponse
//	@Failure	403		{object}	ErrorResponse	"insufficient_scope（write が必要）"
//	@Failure	409		{object}	ErrorResponse	"handle_taken"
//	@Failure	429		{object}	ErrorResponse	"rate_limited"
//	@Router		/v1/profile [put]
func PublicUpdateProfile() {}

// PublicListMatches は対戦履歴を返す。
//
//	@Summary	対戦履歴（Public API）
//	@Tags		public
//	@Produce	json
//	@Security	BearerAuth
//	@Param		from	query	string	false	"この日時以降（RFC3339 か YYYY-MM-DD）"
//	@Param		to		query	string	false	"この日時まで"
//	@Param		mode	query	string	false	"モード"	Enums(ranked, casual, ai, friend)
//	@Param		outcome	query	string	false	"勝敗"	Enums(win, loss, draw)
//	@Param		limit	query	int	false	"件数"	default(20)	minimum(1)	maximum(100)
//	@Param		offset	query	int	false	"開始位置"	default(0)	minimum(0)
//	@Success	200	{object}	MatchListResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	403	{object}	ErrorResponse
//	@Failure	429	{object}	ErrorResponse
//	@Router		/v1/matches [get]
func PublicListMatches() {}

// PublicGetStats は集計値を返す。
//
//	@Summary	統計サマリー（Public API）
//	@Tags		public
//	@Produce	json
//	@Security	BearerAuth
//	@Param		from	query	string	false	"この日時以降"
//	@Param		to		query	string	false	"この日時まで"
//	@Param		mode	query	string	false	"モード"	Enums(ranked, casual, ai, friend)
//	@Success	200	{object}	StatsSummary
//	@Failure	401	{object}	ErrorResponse
//	@Failure	403	{object}	ErrorResponse
//	@Failure	429	{object}	ErrorResponse
//	@Router		/v1/stats [get]
func PublicGetStats() {}

// PublicLeaderboard はランキングを返す。
//
//	@Summary	ランキング（Public API）
//	@Tags		public
//	@Produce	json
//	@Security	BearerAuth
//	@Param		limit	query	int	false	"件数"	default(20)	minimum(1)	maximum(100)
//	@Param		offset	query	int	false	"開始位置"	default(0)	minimum(0)
//	@Success	200	{object}	LeaderboardResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	403	{object}	ErrorResponse
//	@Failure	429	{object}	ErrorResponse
//	@Router		/v1/leaderboard [get]
func PublicLeaderboard() {}

// PublicListFriends は成立済みのフレンド一覧を返す。
//
//	@Summary	フレンド一覧（Public API）
//	@Tags		public
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	FriendshipListResponse
//	@Failure	401	{object}	ErrorResponse
//	@Failure	403	{object}	ErrorResponse
//	@Failure	429	{object}	ErrorResponse
//	@Router		/v1/friends [get]
func PublicListFriends() {}

// PublicCreateFriendRequest はフレンド申請を送る。
//
//	@Summary	フレンド申請（Public API）
//	@Tags		public
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		request	body		FriendRequestCreateRequest	true	"申請先"
//	@Success	201		{object}	FriendshipResponse
//	@Failure	400		{object}	ErrorResponse	"friend_self"
//	@Failure	401		{object}	ErrorResponse
//	@Failure	403		{object}	ErrorResponse	"insufficient_scope / friend_blocked"
//	@Failure	404		{object}	ErrorResponse	"user_not_found"
//	@Failure	409		{object}	ErrorResponse	"friend_already_requested / already_friends"
//	@Failure	429		{object}	ErrorResponse
//	@Router		/v1/friends/requests [post]
func PublicCreateFriendRequest() {}

// PublicRemoveFriend はフレンド解除（pending なら取り下げ）。
//
//	@Summary	フレンド解除（Public API）
//	@Tags		public
//	@Security	BearerAuth
//	@Param		userId	path	string	true	"相手のユーザー ID"	format(uuid)
//	@Success	204
//	@Failure	401	{object}	ErrorResponse
//	@Failure	403	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse	"friend_not_found"
//	@Failure	429	{object}	ErrorResponse
//	@Router		/v1/friends/{userId} [delete]
func PublicRemoveFriend() {}
