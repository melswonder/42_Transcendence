// 実績と SSE の HTTP 入口。
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// EventSubscriber は SSE の受け口を作る。実装は infrastructure の EventHub。
type EventSubscriber interface {
	Subscribe(userID uuid.UUID) (<-chan domain.MatchEvent, func())
}

// sseKeepAlive は無通信のまま接続が切られないよう、定期的にコメント行を送る間隔。
// プロキシによっては数十秒の無通信で切るため。
const sseKeepAlive = 25 * time.Second

type AchievementHandler struct {
	uc          *usecase.AchievementUsecase
	events      EventSubscriber
	currentUser currentUserFunc
}

func NewAchievementHandler(
	uc *usecase.AchievementUsecase, events EventSubscriber, currentUser currentUserFunc,
) *AchievementHandler {
	return &AchievementHandler{uc: uc, events: events, currentUser: currentUser}
}

type achievementResponse struct {
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Progress    int        `json:"progress"`
	Target      int        `json:"target"`
	Unlocked    bool       `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at"`
}

type achievementListResponse struct {
	Items         []achievementResponse `json:"items"`
	UnlockedCount int                   `json:"unlocked_count"`
	TotalCount    int                   `json:"total_count"`
}

type matchEventResponse struct {
	MatchID    string    `json:"match_id"`
	Outcome    string    `json:"outcome"`
	Rating     int       `json:"rating"`
	OccurredAt time.Time `json:"occurred_at"`
}

// List - GET /achievements/me
func (h *AchievementHandler) List(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	achievements, err := h.uc.List(r.Context(), user.ID)
	if err != nil {
		log.Printf("list achievements: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")

		return
	}

	items := make([]achievementResponse, 0, len(achievements))
	unlocked := 0

	for _, a := range achievements {
		if a.Unlocked {
			unlocked++
		}
		items = append(items, achievementResponse{
			Code:        a.Code,
			Name:        a.Name,
			Description: a.Description,
			Category:    a.Category,
			Progress:    a.Progress,
			Target:      a.Target,
			Unlocked:    a.Unlocked,
			UnlockedAt:  a.UnlockedAt,
		})
	}

	writeJSON(w, http.StatusOK, achievementListResponse{
		Items: items, UnlockedCount: unlocked, TotalCount: len(items),
	})
}

// Stream - GET /stats/stream
//
// 自分が参加した対戦が記録されるたびに 1 イベント流す。
// 接続はクライアントが閉じるまで開きっぱなしになる。
func (h *AchievementHandler) Stream(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "streaming unsupported")

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// nginx などが間に入ったときにバッファされないようにする。
	// 溜め込まれると「リアルタイム」でなくなる。
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	events, unsubscribe := h.events.Subscribe(user.ID)
	defer unsubscribe()

	keepAlive := time.NewTicker(sseKeepAlive)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			// クライアントが閉じた。unsubscribe は defer で走る。
			return

		case event, ok := <-events:
			if !ok {
				return
			}
			if err := writeSSEEvent(w, event); err != nil {
				log.Printf("write sse event: %v", err)

				return
			}
			flusher.Flush()

		case <-keepAlive.C:
			// コメント行。クライアントからは無視されるが、接続は生き続ける。
			if _, err := fmt.Fprint(w, ": keep-alive\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, event domain.MatchEvent) error {
	body, err := json.Marshal(matchEventResponse{
		MatchID:    event.MatchID.String(),
		Outcome:    event.Outcome,
		Rating:     event.Rating,
		OccurredAt: event.OccurredAt,
	})
	if err != nil {
		return err
	}

	// SSE の 1 イベントは "event:" と "data:" の行を空行で閉じる形。
	_, err = fmt.Fprintf(w, "event: match_recorded\ndata: %s\n\n", body)

	return err
}
