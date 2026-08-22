// フレンド関係の HTTP 入口。
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

type FriendHandler struct {
	uc          *usecase.FriendUsecase
	currentUser currentUserFunc
}

func NewFriendHandler(uc *usecase.FriendUsecase, currentUser currentUserFunc) *FriendHandler {
	return &FriendHandler{uc: uc, currentUser: currentUser}
}

// friendshipResponse は apispec の FriendshipResponse と対。
// online はタスク要件（フレンドのオンライン状態）のための拡張フィールド。
type friendshipResponse struct {
	User          userPublicResponse `json:"user"`
	Status        string             `json:"status"`
	RequestedByMe bool               `json:"requested_by_me"`
	Online        bool               `json:"online"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

func toFriendshipResponse(f domain.Friendship) friendshipResponse {
	return friendshipResponse{
		User:          toUserPublic(f.Other),
		Status:        f.Status,
		RequestedByMe: f.RequestedByMe,
		Online:        f.Online,
		CreatedAt:     f.CreatedAt,
		UpdatedAt:     f.UpdatedAt,
	}
}

type friendshipListResponse struct {
	Items []friendshipResponse `json:"items"`
	pagination
}

// List - GET /friends
// 成立済みのフレンド一覧（status で pending / rejected も引ける）。
func (h *FriendHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	status := q.Get("status")
	if status == "" {
		status = domain.FriendStatusAccepted
	}
	if status != domain.FriendStatusPending && status != domain.FriendStatusAccepted && status != domain.FriendStatusRejected {
		writeJSONError(w, http.StatusBadRequest, "invalid status")
		return
	}

	h.writeList(w, r, user.ID, usecase.FriendListFilter{
		Status: status,
		Limit:  parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit),
		Offset: parseBoundedInt(q.Get("offset"), 0, 0, 0),
	})
}

// ListRequests - GET /friends/requests
// 未応答の申請一覧。direction で「届いた申請」と「送った申請」を切り替える。
func (h *FriendHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	direction := q.Get("direction")
	if direction == "" {
		direction = "incoming"
	}
	if direction != "incoming" && direction != "outgoing" {
		writeJSONError(w, http.StatusBadRequest, "invalid direction")
		return
	}

	h.writeList(w, r, user.ID, usecase.FriendListFilter{
		Status:    domain.FriendStatusPending,
		Direction: direction,
		Limit:     parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit),
		Offset:    parseBoundedInt(q.Get("offset"), 0, 0, 0),
	})
}

func (h *FriendHandler) writeList(w http.ResponseWriter, r *http.Request, userID uuid.UUID, f usecase.FriendListFilter) {
	friendships, total, err := h.uc.List(r.Context(), userID, f)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list friendships")
		return
	}

	items := make([]friendshipResponse, 0, len(friendships))
	for _, friendship := range friendships {
		items = append(items, toFriendshipResponse(friendship))
	}
	writeJSON(w, http.StatusOK, friendshipListResponse{
		Items:      items,
		pagination: pagination{Total: total, Limit: f.Limit, Offset: f.Offset},
	})
}

type friendRequestCreateRequest struct {
	UserID string `json:"user_id"`
}

// CreateRequest - POST /friends/requests
// 相手が既にこちらへ申請していれば、その場で accepted になる。
func (h *FriendHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req friendRequestCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	otherID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	friendship, err := h.uc.Request(r.Context(), user.ID, otherID)
	writeFriendRequestResult(w, friendship, err)
}

// writeFriendRequestResult は申請の結果をコード付きで書く。Public API とも共用する。
func writeFriendRequestResult(w http.ResponseWriter, friendship *domain.Friendship, err error) {
	switch {
	case err == nil:
		writeJSON(w, http.StatusCreated, toFriendshipResponse(*friendship))
	case errors.Is(err, domain.ErrFriendSelf):
		writeJSONErrorCode(w, http.StatusBadRequest, "friend_self", "cannot befriend yourself")
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSONErrorCode(w, http.StatusNotFound, "user_not_found", "user not found")
	case errors.Is(err, domain.ErrFriendBlocked):
		writeJSONErrorCode(w, http.StatusForbidden, "friend_blocked", "cannot send a request to this user")
	case errors.Is(err, domain.ErrFriendAlreadyRequested):
		writeJSONErrorCode(w, http.StatusConflict, "friend_already_requested", "friend request already sent")
	case errors.Is(err, domain.ErrAlreadyFriends):
		writeJSONErrorCode(w, http.StatusConflict, "already_friends", "already friends")
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to create friend request")
	}
}

type friendRequestDecisionRequest struct {
	Action string `json:"action"` // accept | reject
}

// Decide - PATCH /friends/requests/{userId}
// 届いている申請に応答する。
func (h *FriendHandler) Decide(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	otherID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeJSONErrorCode(w, http.StatusNotFound, "friend_not_found", "friend request not found")
		return
	}

	var req friendRequestDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Action != "accept" && req.Action != "reject" {
		writeJSONError(w, http.StatusBadRequest, "action must be accept or reject")
		return
	}

	friendship, err := h.uc.Decide(r.Context(), user.ID, otherID, req.Action == "accept")
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toFriendshipResponse(*friendship))
	case errors.Is(err, domain.ErrFriendshipNotFound):
		writeJSONErrorCode(w, http.StatusNotFound, "friend_not_found", "friend request not found")
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to update friend request")
	}
}

// Remove - DELETE /friends/{userId}
// 成立済みなら解除、pending なら申請の取り下げ。
func (h *FriendHandler) Remove(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	otherID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeJSONErrorCode(w, http.StatusNotFound, "friend_not_found", "friendship not found")
		return
	}

	err = h.uc.Remove(r.Context(), user.ID, otherID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, domain.ErrFriendshipNotFound):
		writeJSONErrorCode(w, http.StatusNotFound, "friend_not_found", "friendship not found")
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to remove friendship")
	}
}
