package infrastructure

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// PresenceHub はユーザーを最後に見かけた時刻をメモリで覚えている。
//
// 認証付きリクエストと対局中の WebSocket keep-alive のたびに Touch され、
// 「最近見かけたか」でオンライン状態を導出する。DB には書かない
// （毎リクエストの UPDATE を増やさないため）。再起動で消えるが、
// presence は数分で埋め直される揮発情報なので問題ない。
// 複数インスタンス構成にするときは Redis へ移す（README の #3）。
type PresenceHub struct {
	mu   sync.RWMutex
	seen map[uuid.UUID]time.Time
}

func NewPresenceHub() *PresenceHub {
	return &PresenceHub{seen: make(map[uuid.UUID]time.Time)}
}

// Touch はユーザーを見かけたことを記録する。
func (h *PresenceHub) Touch(userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seen[userID] = time.Now()
}

// LastSeen は最後に見かけた時刻。一度も見ていなければ false。
func (h *PresenceHub) LastSeen(userID uuid.UUID) (time.Time, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	t, ok := h.seen[userID]
	return t, ok
}
