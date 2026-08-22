// プロフィールとユーザー検索の HTTP 入口。
package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

type UserHandler struct {
	uc          *usecase.UserUsecase
	currentUser currentUserFunc
}

func NewUserHandler(uc *usecase.UserUsecase, currentUser currentUserFunc) *UserHandler {
	return &UserHandler{uc: uc, currentUser: currentUser}
}

// userPublicResponse は他人にも見せられる範囲。apispec の UserPublic と対。
type userPublicResponse struct {
	ID               string    `json:"id"`
	DisplayName      string    `json:"display_name"`
	Handle           string    `json:"handle"`
	AvatarURL        *string   `json:"avatar_url"` // 未設定なら null。表示側が頭文字アバターに落とす
	Level            int       `json:"level"`
	ExperiencePoints int       `json:"experience_points"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

// userPrivateResponse は自分にだけ見せる範囲。apispec の UserPrivate と対。
type userPrivateResponse struct {
	userPublicResponse
	Email           *string   `json:"email"`
	PreferredLocale string    `json:"preferred_locale"`
	AvatarAssetID   *string   `json:"avatar_asset_id"` // 設定画面が削除・差し替えに使う
	HasPassword     bool      `json:"has_password"`
	LinkedProviders []string  `json:"linked_providers"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// avatarURLFor はアセット ID を配信 URL に変える。実体は GET /media/{id}/file。
func avatarURLFor(assetID *uuid.UUID) *string {
	if assetID == nil {
		return nil
	}
	url := "/media/" + assetID.String() + "/file"
	return &url
}

func toUserPublic(u domain.User) userPublicResponse {
	return userPublicResponse{
		ID:               u.ID.String(),
		DisplayName:      u.DisplayName,
		Handle:           u.Handle,
		AvatarURL:        avatarURLFor(u.AvatarAssetID),
		Level:            u.Level,
		ExperiencePoints: u.ExperiencePoints,
		Status:           u.Status,
		CreatedAt:        u.CreatedAt,
	}
}

func toUserPrivate(p *domain.Profile) userPrivateResponse {
	providers := p.LinkedProviders
	if providers == nil {
		providers = []string{}
	}
	var avatarAssetID *string
	if p.AvatarAssetID != nil {
		id := p.AvatarAssetID.String()
		avatarAssetID = &id
	}
	return userPrivateResponse{
		userPublicResponse: toUserPublic(p.User),
		Email:              p.Email,
		PreferredLocale:    p.PreferredLocale,
		AvatarAssetID:      avatarAssetID,
		HasPassword:        p.HasPassword,
		LinkedProviders:    providers,
		UpdatedAt:          p.UpdatedAt,
	}
}

// Me - GET /users/me
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	profile, err := h.uc.MyProfile(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}
	writeJSON(w, http.StatusOK, toUserPrivate(profile))
}

// updateMeRequest はプロフィール編集の入力。apispec の UpdateMeRequest と対。
// avatar_asset_id は「省略 = 変更しない」「null = 外す」を区別するため RawMessage で受ける。
type updateMeRequest struct {
	DisplayName     *string         `json:"display_name"`
	Handle          *string         `json:"handle"`
	PreferredLocale *string         `json:"preferred_locale"`
	AvatarAssetID   json.RawMessage `json:"avatar_asset_id"`
}

// UpdateMe - PATCH /users/me
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	update := domain.ProfileUpdate{
		DisplayName:     req.DisplayName,
		Handle:          req.Handle,
		PreferredLocale: req.PreferredLocale,
	}
	if len(req.AvatarAssetID) > 0 {
		update.AvatarSet = true
		if string(req.AvatarAssetID) != "null" {
			var raw string
			if err := json.Unmarshal(req.AvatarAssetID, &raw); err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid avatar_asset_id")
				return
			}
			assetID, err := uuid.Parse(raw)
			if err != nil {
				writeJSONError(w, http.StatusBadRequest, "invalid avatar_asset_id")
				return
			}
			update.AvatarAssetID = &assetID
		}
	}

	profile, err := h.uc.UpdateMe(r.Context(), user.ID, update)
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toUserPrivate(profile))
	case errors.Is(err, domain.ErrHandleTaken):
		writeJSONErrorCode(w, http.StatusConflict, "handle_taken", "handle already taken")
	case errors.Is(err, domain.ErrInvalidDisplayName):
		writeJSONErrorCode(w, http.StatusBadRequest, "invalid_display_name", err.Error())
	case errors.Is(err, domain.ErrInvalidHandle):
		writeJSONErrorCode(w, http.StatusBadRequest, "invalid_handle", err.Error())
	case errors.Is(err, domain.ErrInvalidLocale):
		writeJSONErrorCode(w, http.StatusBadRequest, "invalid_locale", err.Error())
	case errors.Is(err, domain.ErrInvalidAvatarAsset):
		writeJSONErrorCode(w, http.StatusBadRequest, "invalid_avatar_asset", err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to update profile")
	}
}

// Get - GET /users/{userId}
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	if _, err := h.currentUser(r); err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	target, err := h.uc.GetPublic(r.Context(), userID)
	if errors.Is(err, domain.ErrUserNotFound) {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load user")
		return
	}
	writeJSON(w, http.StatusOK, toUserPublic(*target))
}

// pagination は一覧レスポンス共通のページ情報。apispec の Pagination と対。
type pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type userListResponse struct {
	Items []userPublicResponse `json:"items"`
	pagination
}

// List - GET /users
// フレンド申請の相手探し用の検索。
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	filter := usecase.UserSearchFilter{
		Query:  q.Get("q"),
		Handle: q.Get("handle"),
		Limit:  parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit),
		Offset: parseBoundedInt(q.Get("offset"), 0, 0, 0),
	}

	users, total, err := h.uc.Search(r.Context(), user.ID, filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to search users")
		return
	}

	items := make([]userPublicResponse, 0, len(users))
	for _, u := range users {
		items = append(items, toUserPublic(u))
	}
	writeJSON(w, http.StatusOK, userListResponse{
		Items:      items,
		pagination: pagination{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	})
}
