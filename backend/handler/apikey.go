// API キー管理の HTTP 入口。ここは Cookie セッションで認証する（キー自身では管理できない）。
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

type APIKeyHandler struct {
	uc          *usecase.APIKeyUsecase
	currentUser currentUserFunc
}

func NewAPIKeyHandler(uc *usecase.APIKeyUsecase, currentUser currentUserFunc) *APIKeyHandler {
	return &APIKeyHandler{uc: uc, currentUser: currentUser}
}

type apiKeyResponse struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"` // どのキーか見分ける用。これでは認証できない
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

// apiKeyCreatedResponse は作成時だけ raw key を含む。二度と取得できない。
type apiKeyCreatedResponse struct {
	APIKey string         `json:"api_key"`
	Key    apiKeyResponse `json:"key"`
}

func toAPIKeyResponse(k domain.APIKey) apiKeyResponse {
	return apiKeyResponse{
		ID:         k.ID.String(),
		Name:       k.Name,
		KeyPrefix:  k.KeyPrefix,
		Scopes:     k.Scopes,
		ExpiresAt:  k.ExpiresAt,
		RevokedAt:  k.RevokedAt,
		LastUsedAt: k.LastUsedAt,
		CreatedAt:  k.CreatedAt,
	}
}

type apiKeyCreateRequest struct {
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes"`
	ExpiresAt *time.Time `json:"expires_at"` // 省略で無期限
}

// Create - POST /apikeys
// raw key はこのレスポンスで一度だけ見せる。DB にはハッシュしか残らない。
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req apiKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	raw, key, err := h.uc.Create(r.Context(), user.ID, req.Name, req.Scopes, req.ExpiresAt)
	if errors.Is(err, domain.ErrInvalidAPIKeyInput) {
		writeJSONErrorCode(w, http.StatusBadRequest, "invalid_api_key_input",
			"name (1-50 chars), scopes (read/write) and a future expires_at are required")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create api key")
		return
	}

	writeJSON(w, http.StatusCreated, apiKeyCreatedResponse{APIKey: raw, Key: toAPIKeyResponse(*key)})
}

type apiKeyListResponse struct {
	Items []apiKeyResponse `json:"items"`
}

// List - GET /apikeys
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	keys, err := h.uc.List(r.Context(), user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list api keys")
		return
	}

	items := make([]apiKeyResponse, 0, len(keys))
	for _, k := range keys {
		items = append(items, toAPIKeyResponse(k))
	}
	writeJSON(w, http.StatusOK, apiKeyListResponse{Items: items})
}

// Revoke - DELETE /apikeys/{keyId}
// 行は消さず失効にする。いつ何が失効されたかを追えるようにしておく。
func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	keyID, err := uuid.Parse(r.PathValue("keyId"))
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "api key not found")
		return
	}

	err = h.uc.Revoke(r.Context(), keyID, user.ID)
	if errors.Is(err, domain.ErrAPIKeyNotFound) {
		writeJSONError(w, http.StatusNotFound, "api key not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to revoke api key")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
