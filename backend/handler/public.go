// Public API（/v1）の HTTP 入口。
//
// Cookie ではなく Authorization: Bearer <api key> で認証する。
// スコープ（read / write）とキー単位のレートリミットをここで通してから、
// 既存の usecase をそのまま呼ぶ。データの正本や検証は内側の層と共通。
package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

type PublicHandler struct {
	keys    *usecase.APIKeyUsecase
	users   *usecase.UserUsecase
	matches *usecase.MatchUsecase
	stats   *usecase.StatsUsecase
	friends *usecase.FriendUsecase
}

func NewPublicHandler(
	keys *usecase.APIKeyUsecase,
	users *usecase.UserUsecase,
	matches *usecase.MatchUsecase,
	stats *usecase.StatsUsecase,
	friends *usecase.FriendUsecase,
) *PublicHandler {
	return &PublicHandler{keys: keys, users: users, matches: matches, stats: stats, friends: friends}
}

type publicHandlerFunc func(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity)

// withKey は Public API 共通の関門。認証 → スコープ → レートリミットの順で通す。
//
// レートリミットの状態は成功時も X-RateLimit-* ヘッダで返し、
// curl だけで「あと何回で 429 になるか」を実演できるようにしておく。
func (h *PublicHandler) withKey(scope string, next publicHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok {
			writeJSONErrorCode(w, http.StatusUnauthorized, "missing_api_key",
				"Authorization: Bearer <api key> is required")
			return
		}

		identity, err := h.keys.Authenticate(r.Context(), strings.TrimSpace(raw))
		switch {
		case errors.Is(err, domain.ErrAPIKeyRevoked):
			writeJSONErrorCode(w, http.StatusUnauthorized, "api_key_revoked", "api key has been revoked")
			return
		case errors.Is(err, domain.ErrAPIKeyExpired):
			writeJSONErrorCode(w, http.StatusUnauthorized, "api_key_expired", "api key has expired")
			return
		case err != nil:
			writeJSONErrorCode(w, http.StatusUnauthorized, "invalid_api_key", "invalid api key")
			return
		}

		limit, remaining, resetAt, err := h.keys.Authorize(identity, scope)
		if errors.Is(err, domain.ErrAPIKeyScope) {
			writeJSONErrorCode(w, http.StatusForbidden, "insufficient_scope",
				"this endpoint requires the '"+scope+"' scope")
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))
		if errors.Is(err, domain.ErrAPIKeyRateLimited) {
			// Retry-After は「あと何秒」で返す（HTTP の仕様に合わせる）。
			w.Header().Set("Retry-After", strconv.Itoa(max(1, int(time.Until(resetAt).Seconds())+1)))
			writeJSONErrorCode(w, http.StatusTooManyRequests, "rate_limited", "rate limit exceeded")
			return
		}

		next(w, r, identity)
	}
}

// Register は /v1 配下の全エンドポイントを繋ぐ。
// CRUD の要件はデータ側のエンドポイント（プロフィール・フレンド・履歴・統計）で満たし、
// キー管理（/apikeys）はここに含めない。
func (h *PublicHandler) Register(router *gin.Engine) {
	v1 := router.Group("/v1")
	{
		v1.GET("/profile", wrapF(h.withKey(domain.APIScopeRead, h.getProfile)))
		v1.PUT("/profile", wrapF(h.withKey(domain.APIScopeWrite, h.updateProfile)))
		v1.GET("/matches", wrapF(h.withKey(domain.APIScopeRead, h.listMatches)))
		v1.GET("/stats", wrapF(h.withKey(domain.APIScopeRead, h.getStats)))
		v1.GET("/leaderboard", wrapF(h.withKey(domain.APIScopeRead, h.getLeaderboard)))
		v1.GET("/friends", wrapF(h.withKey(domain.APIScopeRead, h.listFriends)))
		v1.POST("/friends/requests", wrapF(h.withKey(domain.APIScopeWrite, h.createFriendRequest)))
		v1.DELETE("/friends/:userId", wrapF(h.withKey(domain.APIScopeWrite, h.removeFriend), "userId"))
	}
}

// getProfile - GET /v1/profile（read）
func (h *PublicHandler) getProfile(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
	profile, err := h.users.MyProfile(r.Context(), id.UserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}
	writeJSON(w, http.StatusOK, toUserPrivate(profile))
}

type publicProfileUpdateRequest struct {
	DisplayName     *string `json:"display_name"`
	Handle          *string `json:"handle"`
	PreferredLocale *string `json:"preferred_locale"`
}

// updateProfile - PUT /v1/profile（write）
// アバターの差し替えは含めない（画像のアップロードが Cookie セッション前提のため）。
func (h *PublicHandler) updateProfile(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
	var req publicProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	profile, err := h.users.UpdateMe(r.Context(), id.UserID, domain.ProfileUpdate{
		DisplayName:     req.DisplayName,
		Handle:          req.Handle,
		PreferredLocale: req.PreferredLocale,
	})
	switch {
	case err == nil:
		writeJSON(w, http.StatusOK, toUserPrivate(profile))
	case errors.Is(err, domain.ErrHandleTaken):
		writeJSONErrorCode(w, http.StatusConflict, "handle_taken", "handle already taken")
	case errors.Is(err, domain.ErrInvalidDisplayName),
		errors.Is(err, domain.ErrInvalidHandle),
		errors.Is(err, domain.ErrInvalidLocale):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to update profile")
	}
}

// listMatches - GET /v1/matches（read）
// 絞り込み（from / to / mode / outcome / limit / offset）は画面用 API と同じ。
func (h *PublicHandler) listMatches(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
	filter, err := parseMatchFilter(r, true)
	if err != nil {
		writeMatchError(w, err)
		return
	}

	records, total, err := h.matches.ListMatches(r.Context(), id.UserID, filter)
	if err != nil {
		log.Printf("public list matches: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	items := make([]matchResponse, 0, len(records))
	for _, rec := range records {
		items = append(items, toMatchResponse(rec))
	}
	writeJSON(w, http.StatusOK, matchListResponse{
		Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset,
	})
}

// getStats - GET /v1/stats（read）
func (h *PublicHandler) getStats(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
	filter, err := parseMatchFilter(r, false)
	if err != nil {
		writeMatchError(w, err)
		return
	}

	summary, err := h.stats.Summary(r.Context(), id.UserID, filter)
	if err != nil {
		log.Printf("public stats: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	floor, ceiling := domain.XPRangeForLevel(summary.Level)
	writeJSON(w, http.StatusOK, statsSummaryResponse{
		Wins:           summary.Wins,
		Losses:         summary.Losses,
		Draws:          summary.Draws,
		TotalMatches:   summary.TotalMatches(),
		WinRate:        summary.WinRate(),
		CurrentStreak:  summary.CurrentStreak,
		BestStreak:     summary.BestStreak,
		Rating:         summary.Rating,
		Ranking:        summary.Ranking,
		TotalPlayers:   summary.TotalPlayers,
		Level:          summary.Level,
		XP:             summary.XP,
		XPForLevel:     floor,
		XPForNextLevel: ceiling,
	})
}

// getLeaderboard - GET /v1/leaderboard（read）
func (h *PublicHandler) getLeaderboard(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
	q := r.URL.Query()
	limit := parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit)
	offset := parseBoundedInt(q.Get("offset"), 0, 0, 0)

	entries, me, total, err := h.stats.Leaderboard(r.Context(), id.UserID, limit, offset)
	if err != nil {
		log.Printf("public leaderboard: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	items := make([]leaderboardEntryResponse, 0, len(entries))
	for _, e := range entries {
		items = append(items, toLeaderboardEntry(e))
	}
	res := leaderboardResponse{Items: items, Total: total, Limit: limit, Offset: offset}
	if me != nil {
		entry := toLeaderboardEntry(*me)
		res.Me = &entry
	}
	writeJSON(w, http.StatusOK, res)
}

// listFriends - GET /v1/friends（read）
func (h *PublicHandler) listFriends(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
	q := r.URL.Query()
	filter := usecase.FriendListFilter{
		Status: domain.FriendStatusAccepted,
		Limit:  parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit),
		Offset: parseBoundedInt(q.Get("offset"), 0, 0, 0),
	}

	friendships, total, err := h.friends.List(r.Context(), id.UserID, filter)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to list friends")
		return
	}

	items := make([]friendshipResponse, 0, len(friendships))
	for _, f := range friendships {
		items = append(items, toFriendshipResponse(f))
	}
	writeJSON(w, http.StatusOK, friendshipListResponse{
		Items:      items,
		pagination: pagination{Total: total, Limit: filter.Limit, Offset: filter.Offset},
	})
}

// createFriendRequest - POST /v1/friends/requests（write）
func (h *PublicHandler) createFriendRequest(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
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

	friendship, err := h.friends.Request(r.Context(), id.UserID, otherID)
	writeFriendRequestResult(w, friendship, err)
}

// removeFriend - DELETE /v1/friends/{userId}（write）
func (h *PublicHandler) removeFriend(w http.ResponseWriter, r *http.Request, id *usecase.APIKeyIdentity) {
	otherID, err := uuid.Parse(r.PathValue("userId"))
	if err != nil {
		writeJSONErrorCode(w, http.StatusNotFound, "friend_not_found", "friendship not found")
		return
	}

	err = h.friends.Remove(r.Context(), id.UserID, otherID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, domain.ErrFriendshipNotFound):
		writeJSONErrorCode(w, http.StatusNotFound, "friend_not_found", "friendship not found")
	default:
		writeJSONError(w, http.StatusInternalServerError, "failed to remove friendship")
	}
}
