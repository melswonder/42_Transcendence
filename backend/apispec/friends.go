// friendships テーブル（§3）に対応する公開 API の仕様。
//
// DB では (A,B)/(B,A) の二重登録を防ぐため user_low_id < user_high_id に正規化しているが、
// API では「相手のユーザー ID」だけを見せる。正規化は infrastructure 層の都合なので外に出さない。
package apispec

import "time"

// FriendshipResponse は自分から見た 1 件の友達関係。
type FriendshipResponse struct {
	User          UserPublic `json:"user"` // 相手
	Status        string     `json:"status"          enums:"pending,accepted,rejected" example:"accepted"`
	RequestedByMe bool       `json:"requested_by_me" example:"true"` // requested_by_user_id が自分か
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// FriendshipListResponse は友達・申請の一覧。
type FriendshipListResponse struct {
	Items []FriendshipResponse `json:"items"`
	Pagination
}

// FriendRequestCreateRequest は友達申請。
type FriendRequestCreateRequest struct {
	UserID string `json:"user_id" binding:"required" format:"uuid" example:"3f0c1a2e-9a5b-4c7d-8e1f-2b3c4d5e6f70"`
}

// FriendRequestDecisionRequest は受け取った申請への返答。
type FriendRequestDecisionRequest struct {
	Action string `json:"action" binding:"required" enums:"accept,reject" example:"accept"`
}

// ListFriends godoc
//
//	@Summary		友達一覧
//	@Description	status=accepted の関係を返す。status を指定すると申請中なども取得できる。
//	@Tags			friends
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status	query		string	false	"絞り込む状態"		Enums(pending, accepted, rejected)	default(accepted)
//	@Param			limit	query		int		false	"取得件数 (1-100)"	default(20)							minimum(1)	maximum(100)
//	@Param			offset	query		int		false	"取得開始位置"		default(0)							minimum(0)
//	@Success		200		{object}	FriendshipListResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/friends [get]
func ListFriends() {}

// ListFriendRequests godoc
//
//	@Summary		友達申請の一覧
//	@Description	direction=incoming で自分宛て、outgoing で自分が出した申請を返す。
//	@Tags			friends
//	@Produce		json
//	@Security		BearerAuth
//	@Param			direction	query		string	false	"申請の向き"	Enums(incoming, outgoing)	default(incoming)
//	@Success		200			{object}	FriendshipListResponse
//	@Failure		401			{object}	ErrorResponse
//	@Router			/friends/requests [get]
func ListFriendRequests() {}

// CreateFriendRequest godoc
//
//	@Summary		友達申請を送る
//	@Description	pending の関係を作る。相手が既に自分へ申請していれば、その場で accepted になる。
//	@Tags			friends
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		FriendRequestCreateRequest	true	"申請先"
//	@Success		201		{object}	FriendshipResponse
//	@Failure		400		{object}	ErrorResponse	"自分自身は指定できない"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		403		{object}	ErrorResponse	"どちらかがブロックしている"
//	@Failure		404		{object}	ErrorResponse	"相手が存在しない"
//	@Failure		409		{object}	ErrorResponse	"既に申請済み、または友達"
//	@Router			/friends/requests [post]
func CreateFriendRequest() {}

// RespondFriendRequest godoc
//
//	@Summary		友達申請に返答する
//	@Description	自分宛ての pending な申請を accepted / rejected にする。
//	@Tags			friends
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			userId	path		string							true	"申請してきたユーザーの ID"	format(uuid)
//	@Param			body	body		FriendRequestDecisionRequest	true	"返答"
//	@Success		200		{object}	FriendshipResponse
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse	"自分宛ての申請が無い"
//	@Router			/friends/requests/{userId} [patch]
func RespondFriendRequest() {}

// DeleteFriend godoc
//
//	@Summary		友達を解除／申請を取り消す
//	@Description	関係を削除する。accepted なら友達解除、pending なら申請の取り消しになる。
//	@Tags			friends
//	@Produce		json
//	@Security		BearerAuth
//	@Param			userId	path	string	true	"相手のユーザー ID"	format(uuid)
//	@Success		204		"削除完了"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse	"関係が無い"
//	@Router			/friends/{userId} [delete]
func DeleteFriend() {}
