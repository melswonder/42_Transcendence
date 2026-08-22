package infrastructure

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFixedWindowLimiter(t *testing.T) {
	t.Parallel()

	limiter := NewFixedWindowLimiter(3)
	now := time.Now()
	limiter.now = func() time.Time { return now }
	key := uuid.New()

	// 上限までは通り、remaining が減っていく。
	for i := 0; i < 3; i++ {
		allowed, limit, remaining, _ := limiter.Allow(key)
		if !allowed || limit != 3 || remaining != 2-i {
			t.Fatalf("%d 回目: allowed=%v remaining=%d", i+1, allowed, remaining)
		}
	}
	// 上限を超えたら拒否。
	if allowed, _, _, _ := limiter.Allow(key); allowed {
		t.Error("上限超過は拒否されるはず")
	}
	// 別のキーは巻き込まれない。
	if allowed, _, _, _ := limiter.Allow(uuid.New()); !allowed {
		t.Error("別のキーは独立して数えるはず")
	}
	// 窓が開き直せばまた通る。
	now = now.Add(time.Minute + time.Second)
	if allowed, _, remaining, _ := limiter.Allow(key); !allowed || remaining != 2 {
		t.Errorf("窓が開き直したら通るはず: %v %d", allowed, remaining)
	}
}
