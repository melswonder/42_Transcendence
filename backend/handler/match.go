// 対戦の HTTP 入口。JSON と CSV の入出力だけを担い、判断は usecase に任せる。
package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// errInvalidQuery はクエリの形式そのものが読めないとき。
var errInvalidQuery = errors.New("invalid query parameter")

// currentUserFunc は Cookie からログイン中のユーザーを引く。
// AuthHandler が持っている実装をそのまま渡す。認証の仕組みを 2 箇所に書かないため。
type currentUserFunc func(r *http.Request) (*domain.User, error)

type MatchHandler struct {
	uc          *usecase.MatchUsecase
	currentUser currentUserFunc
}

func NewMatchHandler(uc *usecase.MatchUsecase, currentUser currentUserFunc) *MatchHandler {
	return &MatchHandler{uc: uc, currentUser: currentUser}
}

type matchParticipantInput struct {
	UserID  string `json:"user_id"`
	Seat    int    `json:"seat"`
	Outcome string `json:"outcome"`
}

type matchCreateRequest struct {
	Mode         string                  `json:"mode"`
	ResultType   string                  `json:"result_type"`
	TotalMoves   int                     `json:"total_moves"`
	StartedAt    time.Time               `json:"started_at"`
	FinishedAt   time.Time               `json:"finished_at"`
	Participants []matchParticipantInput `json:"participants"`
}

type userSummary struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Handle      string `json:"handle"`
	Level       int    `json:"level"`
}

type matchResponse struct {
	ID           string      `json:"id"`
	Mode         string      `json:"mode"`
	Opponent     userSummary `json:"opponent"`
	Outcome      string      `json:"outcome"`
	ResultType   string      `json:"result_type"`
	RatingBefore int         `json:"rating_before"`
	RatingAfter  int         `json:"rating_after"`
	RatingDiff   int         `json:"rating_diff"`
	XPGained     int         `json:"xp_gained"`
	TotalMoves   int         `json:"total_moves"`
	StartedAt    time.Time   `json:"started_at"`
	FinishedAt   time.Time   `json:"finished_at"`
}

type matchListResponse struct {
	Items  []matchResponse `json:"items"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// Create - POST /matches
//
// ゲーム側が決着時に呼ぶ入口。レーティングと XP はサーバーで計算するので、
// 受け取るのは「誰が・どの席で・勝ったか負けたか」だけ。
func (h *MatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	var req matchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")

		return
	}

	participants := make([]domain.MatchParticipant, 0, len(req.Participants))
	for _, p := range req.Participants {
		id, err := uuid.Parse(p.UserID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid user_id")

			return
		}
		participants = append(participants, domain.MatchParticipant{
			UserID: id, Seat: p.Seat, Outcome: p.Outcome,
		})
	}

	match, users, err := h.uc.RecordMatch(r.Context(), domain.Match{
		Mode:         req.Mode,
		ResultType:   req.ResultType,
		TotalMoves:   req.TotalMoves,
		StartedAt:    req.StartedAt,
		FinishedAt:   req.FinishedAt,
		Participants: participants,
	})
	if err != nil {
		writeMatchError(w, err)

		return
	}

	writeJSON(w, http.StatusCreated, recordedMatchResponse(match, users, user.ID))
}

// List - GET /matches
func (h *MatchHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	filter, err := parseMatchFilter(r, true)
	if err != nil {
		writeMatchError(w, err)

		return
	}

	records, total, err := h.uc.ListMatches(r.Context(), user.ID, filter)
	if err != nil {
		log.Printf("list matches: %v", err)
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

// ExportCSV - GET /matches/export.csv
//
// 絞り込みは List と同じで、ページングだけ無視して全件返す。
// CSV をサーバーで作るのは、取得済みの 1 ページぶんではなく
// 条件に合う全件を書き出したいため。
func (h *MatchHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	filter, err := parseMatchFilter(r, false)
	if err != nil {
		writeMatchError(w, err)

		return
	}
	filter.Limit = csvExportLimit

	records, _, err := h.uc.ListMatches(r.Context(), user.ID, filter)
	if err != nil {
		log.Printf("export matches: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")

		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="matches.csv"`)

	writeMatchesCSV(w, records)
}

// csvExportLimit は 1 回の書き出しの上限。
// 無制限にすると 1 リクエストでメモリを使い切れてしまうので蓋をする。
const csvExportLimit = 10000

func writeMatchesCSV(w http.ResponseWriter, records []domain.MatchRecord) {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	header := []string{
		"finished_at", "mode", "opponent_handle", "opponent_display_name",
		"outcome", "result_type", "rating_before", "rating_after", "rating_diff",
		"xp_gained", "total_moves",
	}
	if err := cw.Write(header); err != nil {
		log.Printf("write csv header: %v", err)

		return
	}

	for _, rec := range records {
		row := []string{
			rec.FinishedAt.Format(time.RFC3339),
			rec.Mode,
			rec.Opponent.Handle,
			rec.Opponent.DisplayName,
			rec.Outcome,
			rec.ResultType,
			strconv.Itoa(rec.RatingBefore),
			strconv.Itoa(rec.RatingAfter),
			strconv.Itoa(rec.RatingAfter - rec.RatingBefore),
			strconv.Itoa(rec.XPGained),
			strconv.Itoa(rec.TotalMoves),
		}
		if err := cw.Write(row); err != nil {
			// ヘッダーは送信済みなのでステータスは変えられない。ログにだけ残す。
			log.Printf("write csv row: %v", err)

			return
		}
	}
}

func toUserSummary(u domain.User) userSummary {
	return userSummary{
		ID:          u.ID.String(),
		DisplayName: u.DisplayName,
		Handle:      u.Handle,
		Level:       u.Level,
	}
}

func toMatchResponse(rec domain.MatchRecord) matchResponse {
	return matchResponse{
		ID:           rec.ID.String(),
		Mode:         rec.Mode,
		Opponent:     toUserSummary(rec.Opponent),
		Outcome:      rec.Outcome,
		ResultType:   rec.ResultType,
		RatingBefore: rec.RatingBefore,
		RatingAfter:  rec.RatingAfter,
		RatingDiff:   rec.RatingAfter - rec.RatingBefore,
		XPGained:     rec.XPGained,
		TotalMoves:   rec.TotalMoves,
		StartedAt:    rec.StartedAt,
		FinishedAt:   rec.FinishedAt,
	}
}

// recordedMatchResponse は記録直後の対戦を、呼び出したユーザーから見た形にする。
//
// 呼び出した本人が参加していないこともある（観戦サーバーが記録する場合など）。
// その場合は自分の結果が空のまま、相手として先頭の参加者を入れて返す。
func recordedMatchResponse(
	match *domain.Match, users map[uuid.UUID]domain.User, viewerID uuid.UUID,
) matchResponse {
	res := matchResponse{
		ID:         match.ID.String(),
		Mode:       match.Mode,
		ResultType: match.ResultType,
		TotalMoves: match.TotalMoves,
		StartedAt:  match.StartedAt,
		FinishedAt: match.FinishedAt,
	}

	for _, p := range match.Participants {
		switch p.UserID {
		case viewerID:
			res.Outcome = p.Outcome
			res.RatingBefore = p.RatingBefore
			res.RatingAfter = p.RatingAfter
			res.RatingDiff = p.RatingAfter - p.RatingBefore
			res.XPGained = p.XPGained
		default:
			res.Opponent = toUserSummary(users[p.UserID])
		}
	}

	return res
}

// writeMatchError はドメインの失敗を HTTP のステータスへ翻訳する。
// 内訳は 400 に寄せる。どれもクライアントが送った内容の問題なので。
func writeMatchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		writeJSONError(w, http.StatusNotFound, "participant not found")
	case errors.Is(err, domain.ErrInvalidMatchMode),
		errors.Is(err, domain.ErrInvalidResultType),
		errors.Is(err, domain.ErrInvalidMatchPeriod),
		errors.Is(err, domain.ErrInvalidParticipants),
		errors.Is(err, domain.ErrInvalidOutcome),
		errors.Is(err, domain.ErrInconsistentOutcome),
		errors.Is(err, domain.ErrInvalidDateRange),
		errors.Is(err, errInvalidQuery):
		writeJSONError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("match: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
	}
}
