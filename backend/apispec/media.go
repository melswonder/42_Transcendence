// media_assets テーブル（§3）に対応する公開 API の仕様。
//
// storage_key は「URL からの列挙を防ぐ推測困難な値」なので API では返さず、
// 代わりに配信用の URL だけを見せる。
package apispec

import "time"

// MediaAssetResponse はアップロード済みアセット 1 件。
type MediaAssetResponse struct {
	ID               string    `json:"id"                format:"uuid"`
	Purpose          string    `json:"purpose"           enums:"avatar" example:"avatar"`
	URL              string    `json:"url"               example:"https://cdn.example.com/avatars/9b7f2c11.png"`
	OriginalFilename string    `json:"original_filename" example:"icon.png"`
	MimeType         string    `json:"mime_type"         enums:"image/png,image/jpeg,image/webp" example:"image/png"`
	SizeBytes        int64     `json:"size_bytes"        example:"48219"`
	Width            *int      `json:"width"             example:"512"`
	Height           *int      `json:"height"            example:"512"`
	ChecksumSHA256   string    `json:"checksum_sha256"   example:"9f2b7c1e4a5d6f708192a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f8"`
	Status           string    `json:"status"            enums:"active,deleted" example:"active"`
	CreatedAt        time.Time `json:"created_at"`
}

// MediaAssetListResponse は自分のアセット一覧。
type MediaAssetListResponse struct {
	Items []MediaAssetResponse `json:"items"`
	Pagination
}

// UploadAvatar godoc
//
//	@Summary		アバター画像をアップロード
//	@Description	multipart/form-data で画像を送る。アップロードしただけでは適用されず、PATCH /users/me の avatar_asset_id で設定する。
//	@Tags			media
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			file	formData	file	true	"画像ファイル（png / jpeg / webp、最大 5MB）"
//	@Success		201		{object}	MediaAssetResponse
//	@Failure		400		{object}	ErrorResponse	"ファイルが無い、または壊れている"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		413		{object}	ErrorResponse	"サイズ上限を超えている"
//	@Failure		415		{object}	ErrorResponse	"未対応の形式"
//	@Router			/media/avatars [post]
func UploadAvatar() {}

// ListMedia godoc
//
//	@Summary		自分のアセット一覧
//	@Description	status=active のアセットを新しい順に返す。
//	@Tags			media
//	@Produce		json
//	@Security		BearerAuth
//	@Param			purpose	query		string	false	"用途で絞り込む"		Enums(avatar)
//	@Param			limit	query		int		false	"取得件数 (1-100)"	default(20)	minimum(1)	maximum(100)
//	@Param			offset	query		int		false	"取得開始位置"		default(0)	minimum(0)
//	@Success		200		{object}	MediaAssetListResponse
//	@Failure		401		{object}	ErrorResponse
//	@Router			/media [get]
func ListMedia() {}

// GetMedia godoc
//
//	@Summary	アセットのメタ情報を取得
//	@Tags		media
//	@Produce	json
//	@Security	BearerAuth
//	@Param		assetId	path		string	true	"アセット ID"	format(uuid)
//	@Success	200		{object}	MediaAssetResponse
//	@Failure	401		{object}	ErrorResponse
//	@Failure	404		{object}	ErrorResponse
//	@Router		/media/{assetId} [get]
func GetMedia() {}

// DeleteMedia godoc
//
//	@Summary		アセットを削除
//	@Description	論理削除（status=deleted, deleted_at）。使用中のアバターを消すとデフォルト画像に戻る。
//	@Tags			media
//	@Produce		json
//	@Security		BearerAuth
//	@Param			assetId	path	string	true	"アセット ID"	format(uuid)
//	@Success		204		"削除完了"
//	@Failure		401		{object}	ErrorResponse
//	@Failure		404		{object}	ErrorResponse	"自分のアセットに存在しない"
//	@Router			/media/{assetId} [delete]
func DeleteMedia() {}
