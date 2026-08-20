package infrastructure

import (
	"sync"

	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// eventBuffer は 1 接続あたりに溜められるイベント数。
//
// 読み手が遅いときにここが埋まったら、そのイベントは捨てる。
// ブロックすると対戦の記録そのものが止まってしまうため、
// 通知の取りこぼしより記録の完了を優先する。
const eventBuffer = 8

// EventHub は「誰に何が起きたか」をプロセス内で配る。
//
// 複数インスタンスで動かすときは Redis の Pub/Sub などに置き換える必要がある。
// 今は 1 プロセスなので、その手前の一番単純な形にしてある。
type EventHub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID]map[chan domain.MatchEvent]struct{}
}

var _ usecase.MatchNotifier = (*EventHub)(nil)

func NewEventHub() *EventHub {
	return &EventHub{subscribers: map[uuid.UUID]map[chan domain.MatchEvent]struct{}{}}
}

// Subscribe は 1 接続ぶんの受け口を作る。
// 返り値の関数を必ず呼んで後始末すること（呼ばないと購読が残り続ける）。
func (h *EventHub) Subscribe(userID uuid.UUID) (<-chan domain.MatchEvent, func()) {
	ch := make(chan domain.MatchEvent, eventBuffer)

	h.mu.Lock()
	if h.subscribers[userID] == nil {
		h.subscribers[userID] = map[chan domain.MatchEvent]struct{}{}
	}
	h.subscribers[userID][ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() { h.unsubscribe(userID, ch) }
}

func (h *EventHub) unsubscribe(userID uuid.UUID, ch chan domain.MatchEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	channels, ok := h.subscribers[userID]
	if !ok {
		return
	}

	delete(channels, ch)
	// 最後の 1 本が消えたらキーごと落とす。放っておくと map が増え続ける。
	if len(channels) == 0 {
		delete(h.subscribers, userID)
	}

	close(ch)
}

// NotifyMatchRecorded は対戦の参加者全員へイベントを配る。
//
// 同じ人が複数のタブを開いていれば、そのぶん配られる。
func (h *EventHub) NotifyMatchRecorded(match domain.Match) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, p := range match.Participants {
		event := domain.MatchEvent{
			MatchID:    match.ID,
			UserID:     p.UserID,
			Outcome:    p.Outcome,
			Rating:     p.RatingAfter,
			OccurredAt: match.FinishedAt,
		}

		for ch := range h.subscribers[p.UserID] {
			select {
			case ch <- event:
			default:
				// 受け口が詰まっている。取りこぼすが、記録は止めない。
			}
		}
	}
}
