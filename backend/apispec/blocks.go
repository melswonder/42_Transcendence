// blocks テーブル（§3）に対応する公開 API の仕様。
// friendships と違い方向を持つ関係なので、「自分がブロックした相手」だけを一覧で返す。
package apispec

import "time"

// BlockResponse は自分がブロックしている 1 件。
type BlockResponse struct {
	ID        string     `json:"id"   format:"uuid"`
	User      UserPublic `json:"user"` // ブロックされている側
	CreatedAt time.Time  `json:"created_at"`
}

// BlockListResponse はブロック一覧。
type BlockListResponse struct {
	Items []BlockResponse `json:"items"`
	Pagination
}

// BlockCreateRequest はブロックの追加。
type BlockCreateRequest struct {
	UserID string `json:"user_id" binding:"required" format:"uuid" example:"3f0c1a2e-9a5b-4c7d-8e1f-2b3c4d5e6f70"`
}

// ListBlocks godoc
//
//	@Summary		ブロック一覧
//	@Description	自分がブロックしている相手を返す。自分をブロックしている相手は取得できない。
//	@Tags			blocks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int	false	"取得件数 (1-100)"	default(20)	minimum(1)	maximum(100)
//	@Param			offset	query		int	false	"取得開始位置"		default(0)	minimum(0)
//	@Success		200		{object}	BlockListResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/blocks [get]
func ListBlocks() {}

// CreateBlock godoc
//
//	@Summary		ユーザーをブロック
//	@Description	既存の友達関係があれば同時に解除する。同じ相手を二重にブロックはできない。
//	@Tags			blocks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		BlockCreateRequest	true	"ブロックする相手"
//	@Success		201		{object}	BlockResponse
//	@Failure		400		{object}	ErrorResponse	"自分自身は指定できない"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse	"相手が存在しない"
//	@Failure		409		{object}	ErrorResponse	"既にブロック済み"
//	@Router			/blocks [post]
func CreateBlock() {}

// DeleteBlock godoc
//
//	@Summary	ブロックを解除
//	@Tags		blocks
//	@Produce	json
//	@Security	BearerAuth
//	@Param		userId	path	string	true	"ブロックを解除する相手のユーザー ID"	format(uuid)
//	@Success	204		"解除完了"
//	@Failure	401		{object}	ErrorResponse
//	@Failure	404		{object}	ErrorResponse	"ブロックしていない"
//	@Router		/blocks/{userId} [delete]
func DeleteBlock() {}
