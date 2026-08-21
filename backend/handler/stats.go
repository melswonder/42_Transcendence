// 統計の HTTP 入口。集計は usecase に任せ、ここは JSON への詰め替えだけを行う。
package handler

import (
	"log"
	"net/http"
	"time"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

type StatsHandler struct {
	uc          *usecase.StatsUsecase
	currentUser currentUserFunc
}

func NewStatsHandler(uc *usecase.StatsUsecase, currentUser currentUserFunc) *StatsHandler {
	return &StatsHandler{uc: uc, currentUser: currentUser}
}

type statsSummaryResponse struct {
	Wins           int     `json:"wins"`
	Losses         int     `json:"losses"`
	Draws          int     `json:"draws"`
	TotalMatches   int     `json:"total_matches"`
	WinRate        float64 `json:"win_rate"`
	CurrentStreak  int     `json:"current_streak"`
	BestStreak     int     `json:"best_streak"`
	Rating         int     `json:"rating"`
	Ranking        int     `json:"ranking"`
	TotalPlayers   int     `json:"total_players"`
	Level          int     `json:"level"`
	XP             int     `json:"xp"`
	XPForLevel     int     `json:"xp_for_level"`      // このレベルの下限。進捗バーの起点
	XPForNextLevel int     `json:"xp_for_next_level"` // 次のレベルまでの累計。進捗バーの分母
}

type timeseriesPointResponse struct {
	Date    string `json:"date"`
	Wins    int    `json:"wins"`
	Losses  int    `json:"losses"`
	Draws   int    `json:"draws"`
	Matches int    `json:"matches"`
	Rating  int    `json:"rating"`
}

type timeseriesResponse struct {
	Interval string                    `json:"interval"`
	Points   []timeseriesPointResponse `json:"points"`
}

type breakdownSliceResponse struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type breakdownResponse struct {
	ByResultType []breakdownSliceResponse `json:"by_result_type"`
	ByMode       []breakdownSliceResponse `json:"by_mode"`
	ByOutcome    []breakdownSliceResponse `json:"by_outcome"`
}

type leaderboardEntryResponse struct {
	Rank    int         `json:"rank"`
	User    userSummary `json:"user"`
	Rating  int         `json:"rating"`
	Wins    int         `json:"wins"`
	Losses  int         `json:"losses"`
	WinRate float64     `json:"win_rate"`
}

type leaderboardResponse struct {
	Items  []leaderboardEntryResponse `json:"items"`
	Me     *leaderboardEntryResponse  `json:"me"`
	Total  int                        `json:"total"`
	Limit  int                        `json:"limit"`
	Offset int                        `json:"offset"`
}

// Summary - GET /stats/me
func (h *StatsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	user, filter, ok := h.authorizeAndParse(w, r)
	if !ok {
		return
	}

	summary, err := h.uc.Summary(r.Context(), user.ID, filter)
	if err != nil {
		log.Printf("stats summary: %v", err)
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

// Timeseries - GET /stats/me/timeseries
func (h *StatsHandler) Timeseries(w http.ResponseWriter, r *http.Request) {
	user, filter, ok := h.authorizeAndParse(w, r)
	if !ok {
		return
	}

	interval := domain.NormalizeInterval(r.URL.Query().Get("interval"))

	points, err := h.uc.Timeseries(r.Context(), user.ID, filter, interval)
	if err != nil {
		log.Printf("stats timeseries: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")

		return
	}

	items := make([]timeseriesPointResponse, 0, len(points))
	for _, p := range points {
		items = append(items, timeseriesPointResponse{
			// 日付だけを返す。時刻まで返してもグラフの軸には使わないため。
			Date:    p.Date.Format(time.DateOnly),
			Wins:    p.Wins,
			Losses:  p.Losses,
			Draws:   p.Draws,
			Matches: p.Matches,
			Rating:  p.Rating,
		})
	}

	writeJSON(w, http.StatusOK, timeseriesResponse{Interval: interval, Points: items})
}

// Breakdown - GET /stats/me/breakdown
func (h *StatsHandler) Breakdown(w http.ResponseWriter, r *http.Request) {
	user, filter, ok := h.authorizeAndParse(w, r)
	if !ok {
		return
	}

	breakdown, err := h.uc.Breakdown(r.Context(), user.ID, filter)
	if err != nil {
		log.Printf("stats breakdown: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")

		return
	}

	writeJSON(w, http.StatusOK, breakdownResponse{
		ByResultType: toSliceResponses(breakdown.ByResultType),
		ByMode:       toSliceResponses(breakdown.ByMode),
		ByOutcome:    toSliceResponses(breakdown.ByOutcome),
	})
}

// Leaderboard - GET /leaderboard
func (h *StatsHandler) Leaderboard(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	q := r.URL.Query()
	limit := parseBoundedInt(q.Get("limit"), defaultLimit, 1, maxLimit)
	offset := parseBoundedInt(q.Get("offset"), 0, 0, 0)

	entries, me, total, err := h.uc.Leaderboard(r.Context(), user.ID, limit, offset)
	if err != nil {
		log.Printf("leaderboard: %v", err)
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

// authorizeAndParse は各ハンドラの頭で必ず行う「認証 → クエリの検証」をまとめる。
// ok が false のときはレスポンスを書き終えているので、呼び出し側はそのまま return する。
func (h *StatsHandler) authorizeAndParse(
	w http.ResponseWriter, r *http.Request,
) (*domain.User, usecase.MatchFilter, bool) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return nil, usecase.MatchFilter{}, false
	}

	filter, err := parseMatchFilter(r, false)
	if err != nil {
		writeMatchError(w, err)

		return nil, usecase.MatchFilter{}, false
	}

	return user, filter, true
}

func toSliceResponses(slices []domain.BreakdownSlice) []breakdownSliceResponse {
	out := make([]breakdownSliceResponse, 0, len(slices))
	for _, s := range slices {
		out = append(out, breakdownSliceResponse{Key: s.Key, Count: s.Count})
	}

	return out
}

func toLeaderboardEntry(e domain.LeaderboardEntry) leaderboardEntryResponse {
	return leaderboardEntryResponse{
		Rank:    e.Rank,
		User:    toUserSummary(e.User),
		Rating:  e.Rating,
		Wins:    e.Wins,
		Losses:  e.Losses,
		WinRate: e.WinRate(),
	}
}
