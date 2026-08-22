// アバター画像のアップロードと配信の HTTP 入口。
package handler

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// multipart のヘッダーぶんの余白。本体の上限は domain.AvatarMaxBytes。
const avatarUploadOverhead = 64 << 10

type MediaHandler struct {
	uc          *usecase.MediaUsecase
	currentUser currentUserFunc
}

func NewMediaHandler(uc *usecase.MediaUsecase, currentUser currentUserFunc) *MediaHandler {
	return &MediaHandler{uc: uc, currentUser: currentUser}
}

// mediaAssetResponse は apispec の MediaAssetResponse と対。storage_key は絶対に出さない。
type mediaAssetResponse struct {
	ID               string    `json:"id"`
	Purpose          string    `json:"purpose"`
	URL              string    `json:"url"`
	OriginalFilename string    `json:"original_filename"`
	MimeType         string    `json:"mime_type"`
	SizeBytes        int64     `json:"size_bytes"`
	Width            *int      `json:"width"`
	Height           *int      `json:"height"`
	ChecksumSHA256   string    `json:"checksum_sha256"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

func toMediaAssetResponse(a *domain.MediaAsset) mediaAssetResponse {
	return mediaAssetResponse{
		ID:               a.ID.String(),
		Purpose:          a.Purpose,
		URL:              "/media/" + a.ID.String() + "/file",
		OriginalFilename: a.OriginalFilename,
		MimeType:         a.MimeType,
		SizeBytes:        a.SizeBytes,
		Width:            a.Width,
		Height:           a.Height,
		ChecksumSHA256:   a.ChecksumSHA256,
		Status:           a.Status,
		CreatedAt:        a.CreatedAt,
	}
}

// UploadAvatar - POST /media/avatars
//
// multipart/form-data の "file" フィールドで受ける。
// アップロードしただけでは適用されず、PATCH /users/me で avatar_asset_id を指すと反映される。
func (h *MediaHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// 第一関門。超えた時点で読み込みが打ち切られる。
	r.Body = http.MaxBytesReader(w, r.Body, domain.AvatarMaxBytes+avatarUploadOverhead)

	file, header, err := r.FormFile("file")
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "file too large (max 5MB)")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "form field 'file' is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read file")
		return
	}

	asset, err := h.uc.UploadAvatar(r.Context(), user.ID, header.Filename, data)
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toMediaAssetResponse(asset))
	case errors.Is(err, domain.ErrMediaTooLarge):
		writeJSONError(w, http.StatusRequestEntityTooLarge, "file too large (max 5MB)")
	case errors.Is(err, domain.ErrUnsupportedMediaType):
		writeJSONError(w, http.StatusUnsupportedMediaType, "unsupported image type (png / jpeg / webp)")
	case errors.Is(err, domain.ErrInvalidImage):
		writeJSONError(w, http.StatusBadRequest, "file is not a valid image")
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to upload avatar")
	}
}

type mediaListResponse struct {
	Items []mediaAssetResponse `json:"items"`
	pagination
}

// List - GET /media
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	limit := parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit)
	offset := parseBoundedInt(q.Get("offset"), 0, 0, 0)

	assets, total, err := h.uc.List(r.Context(), user.ID, q.Get("purpose"), limit, offset)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list media")
		return
	}

	items := make([]mediaAssetResponse, 0, len(assets))
	for i := range assets {
		items = append(items, toMediaAssetResponse(&assets[i]))
	}
	writeJSON(w, http.StatusOK, mediaListResponse{
		Items:      items,
		pagination: pagination{Total: total, Limit: limit, Offset: offset},
	})
}

// Get - GET /media/{assetId}
func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	assetID, err := uuid.Parse(r.PathValue("assetId"))
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "media asset not found")
		return
	}

	asset, err := h.uc.GetOwned(r.Context(), assetID, user.ID)
	if errors.Is(err, domain.ErrMediaNotFound) {
		writeJSONError(w, http.StatusNotFound, "media asset not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load media")
		return
	}
	writeJSON(w, http.StatusOK, toMediaAssetResponse(asset))
}

// Delete - DELETE /media/{assetId}
// 論理削除。使用中のアバターを消すとデフォルトアバターに戻る。
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	assetID, err := uuid.Parse(r.PathValue("assetId"))
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "media asset not found")
		return
	}

	err = h.uc.Delete(r.Context(), assetID, user.ID)
	if errors.Is(err, domain.ErrMediaNotFound) {
		writeJSONError(w, http.StatusNotFound, "media asset not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to delete media")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// File - GET /media/{assetId}/file
//
// 画像本体の配信。アバターは他人のプロフィールにも表示されるので、
// 認証は要求せず「active なアセットだけ」を条件にする。
// storage key は推測できないランダム値で、URL には ID しか出ない。
func (h *MediaHandler) File(w http.ResponseWriter, r *http.Request) {
	assetID, err := uuid.Parse(r.PathValue("assetId"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	asset, file, err := h.uc.OpenAsset(r.Context(), assetID)
	if errors.Is(err, domain.ErrMediaNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to open media")
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", asset.MimeType)
	// 内容は不変（差し替えは別 ID になる）ので、長めにキャッシュさせてよい。
	w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	http.ServeContent(w, r, "", asset.CreatedAt, file)
}
