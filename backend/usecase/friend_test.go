package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// fakeFriendRepo はメモリ上のフレンド関係。low/high の正規化も再現する。
type fakeFriendRepo struct {
	mu      sync.Mutex
	pairs   map[[2]uuid.UUID]*fakePair
	users   map[uuid.UUID]domain.User
	blocked map[[2]uuid.UUID]bool
}

type fakePair struct {
	status      string
	requestedBy uuid.UUID
	createdAt   time.Time
}

func pairKey(a, b uuid.UUID) [2]uuid.UUID {
	if a.String() < b.String() {
		return [2]uuid.UUID{a, b}
	}
	return [2]uuid.UUID{b, a}
}

func newFakeFriendRepo(users ...domain.User) *fakeFriendRepo {
	r := &fakeFriendRepo{
		pairs:   make(map[[2]uuid.UUID]*fakePair),
		users:   make(map[uuid.UUID]domain.User),
		blocked: make(map[[2]uuid.UUID]bool),
	}
	for _, u := range users {
		r.users[u.ID] = u
	}
	return r
}

func (r *fakeFriendRepo) GetPair(_ context.Context, me, other uuid.UUID) (*FriendPair, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.pairs[pairKey(me, other)]
	if !ok {
		return nil, domain.ErrFriendshipNotFound
	}
	return &FriendPair{Status: p.status, RequestedByMe: p.requestedBy == me}, nil
}

func (r *fakeFriendRepo) InsertPending(_ context.Context, from, to uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pairKey(from, to)
	if _, ok := r.pairs[key]; ok {
		return domain.ErrFriendAlreadyRequested
	}
	r.pairs[key] = &fakePair{status: domain.FriendStatusPending, requestedBy: from, createdAt: time.Now()}
	return nil
}

func (r *fakeFriendRepo) UpdateStatus(_ context.Context, me, other uuid.UUID, status string, requestedBy uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.pairs[pairKey(me, other)]
	if !ok {
		return domain.ErrFriendshipNotFound
	}
	p.status = status
	p.requestedBy = requestedBy
	return nil
}

func (r *fakeFriendRepo) DeletePair(_ context.Context, me, other uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := pairKey(me, other)
	if _, ok := r.pairs[key]; !ok {
		return domain.ErrFriendshipNotFound
	}
	delete(r.pairs, key)
	return nil
}

func (r *fakeFriendRepo) GetFriendship(_ context.Context, me, other uuid.UUID) (*domain.Friendship, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.pairs[pairKey(me, other)]
	if !ok {
		return nil, domain.ErrFriendshipNotFound
	}
	return &domain.Friendship{
		Other:         r.users[other],
		Status:        p.status,
		RequestedByMe: p.requestedBy == me,
		CreatedAt:     p.createdAt,
	}, nil
}

func (r *fakeFriendRepo) ListFriendships(_ context.Context, me uuid.UUID, f FriendListFilter) ([]domain.Friendship, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []domain.Friendship
	for key, p := range r.pairs {
		if key[0] != me && key[1] != me {
			continue
		}
		if f.Status != "" && p.status != f.Status {
			continue
		}
		if f.Direction == "incoming" && p.requestedBy == me {
			continue
		}
		if f.Direction == "outgoing" && p.requestedBy != me {
			continue
		}
		other := key[0]
		if other == me {
			other = key[1]
		}
		out = append(out, domain.Friendship{
			Other:         r.users[other],
			Status:        p.status,
			RequestedByMe: p.requestedBy == me,
		})
	}
	return out, len(out), nil
}

func (r *fakeFriendRepo) HasBlockRelation(_ context.Context, a, b uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.blocked[pairKey(a, b)], nil
}

func (r *fakeFriendRepo) UserExists(_ context.Context, userID uuid.UUID) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.users[userID]
	return ok, nil
}

// fakePresence は固定の「最後に見かけた時刻」を返す。
type fakePresence struct {
	seen map[uuid.UUID]time.Time
}

func (p *fakePresence) Touch(uuid.UUID) {}
func (p *fakePresence) LastSeen(userID uuid.UUID) (time.Time, bool) {
	t, ok := p.seen[userID]
	return t, ok
}

func friendTestUsers() (domain.User, domain.User) {
	return domain.User{ID: uuid.New(), Handle: "alice"}, domain.User{ID: uuid.New(), Handle: "bob"}
}

func TestFriendRequestFlow(t *testing.T) {
	t.Parallel()

	alice, bob := friendTestUsers()
	repo := newFakeFriendRepo(alice, bob)
	uc := NewFriendUsecase(repo, nil)
	ctx := context.Background()

	// alice → bob へ申請。
	friendship, err := uc.Request(ctx, alice.ID, bob.ID)
	if err != nil {
		t.Fatalf("申請できるはず: %v", err)
	}
	if friendship.Status != domain.FriendStatusPending || !friendship.RequestedByMe {
		t.Errorf("自分発の pending になるはず: %+v", friendship)
	}

	// 同じ申請の重複は 409 相当。
	if _, err := uc.Request(ctx, alice.ID, bob.ID); !errors.Is(err, domain.ErrFriendAlreadyRequested) {
		t.Errorf("重複申請は拒否されるはず: %v", err)
	}

	// bob が承認して成立。
	accepted, err := uc.Decide(ctx, bob.ID, alice.ID, true)
	if err != nil || accepted.Status != domain.FriendStatusAccepted {
		t.Fatalf("承認で成立するはず: %v %+v", err, accepted)
	}

	// 成立後の再申請は already friends。
	if _, err := uc.Request(ctx, alice.ID, bob.ID); !errors.Is(err, domain.ErrAlreadyFriends) {
		t.Errorf("成立済みへの申請は拒否されるはず: %v", err)
	}

	// 解除できる。
	if err := uc.Remove(ctx, alice.ID, bob.ID); err != nil {
		t.Fatalf("解除できるはず: %v", err)
	}
	if _, err := repo.GetPair(ctx, alice.ID, bob.ID); !errors.Is(err, domain.ErrFriendshipNotFound) {
		t.Error("解除後は関係が消えるはず")
	}
}

func TestFriendRequestRules(t *testing.T) {
	t.Parallel()

	alice, bob := friendTestUsers()
	ctx := context.Background()

	t.Run("自分自身への申請は拒否", func(t *testing.T) {
		t.Parallel()
		uc := NewFriendUsecase(newFakeFriendRepo(alice, bob), nil)
		if _, err := uc.Request(ctx, alice.ID, alice.ID); !errors.Is(err, domain.ErrFriendSelf) {
			t.Errorf("got %v", err)
		}
	})

	t.Run("実在しない相手は 404", func(t *testing.T) {
		t.Parallel()
		uc := NewFriendUsecase(newFakeFriendRepo(alice), nil)
		if _, err := uc.Request(ctx, alice.ID, uuid.New()); !errors.Is(err, domain.ErrUserNotFound) {
			t.Errorf("got %v", err)
		}
	})

	t.Run("ブロック関係があれば 403", func(t *testing.T) {
		t.Parallel()
		repo := newFakeFriendRepo(alice, bob)
		repo.blocked[pairKey(alice.ID, bob.ID)] = true
		uc := NewFriendUsecase(repo, nil)
		if _, err := uc.Request(ctx, alice.ID, bob.ID); !errors.Is(err, domain.ErrFriendBlocked) {
			t.Errorf("got %v", err)
		}
	})

	t.Run("相互申請は自動で成立する", func(t *testing.T) {
		t.Parallel()
		repo := newFakeFriendRepo(alice, bob)
		uc := NewFriendUsecase(repo, nil)
		if _, err := uc.Request(ctx, alice.ID, bob.ID); err != nil {
			t.Fatal(err)
		}
		friendship, err := uc.Request(ctx, bob.ID, alice.ID)
		if err != nil {
			t.Fatalf("相互申請はエラーにならないはず: %v", err)
		}
		if friendship.Status != domain.FriendStatusAccepted {
			t.Errorf("その場で accepted になるはず: %+v", friendship)
		}
	})

	t.Run("拒否した後に再申請できる", func(t *testing.T) {
		t.Parallel()
		repo := newFakeFriendRepo(alice, bob)
		uc := NewFriendUsecase(repo, nil)
		if _, err := uc.Request(ctx, alice.ID, bob.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := uc.Decide(ctx, bob.ID, alice.ID, false); err != nil {
			t.Fatal(err)
		}
		friendship, err := uc.Request(ctx, bob.ID, alice.ID)
		if err != nil || friendship.Status != domain.FriendStatusPending {
			t.Errorf("拒否後の再申請は pending に戻るはず: %v %+v", err, friendship)
		}
	})

	t.Run("自分が送った申請は自分で承認できない", func(t *testing.T) {
		t.Parallel()
		repo := newFakeFriendRepo(alice, bob)
		uc := NewFriendUsecase(repo, nil)
		if _, err := uc.Request(ctx, alice.ID, bob.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := uc.Decide(ctx, alice.ID, bob.ID, true); !errors.Is(err, domain.ErrFriendshipNotFound) {
			t.Errorf("自作自演の承認は 404 のはず: %v", err)
		}
	})
}

func TestFriendOnlineStatus(t *testing.T) {
	t.Parallel()

	alice, bob := friendTestUsers()
	repo := newFakeFriendRepo(alice, bob)
	ctx := context.Background()

	// alice と bob は成立済み。
	if err := repo.InsertPending(ctx, alice.ID, bob.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateStatus(ctx, alice.ID, bob.ID, domain.FriendStatusAccepted, bob.ID); err != nil {
		t.Fatal(err)
	}

	presence := &fakePresence{seen: map[uuid.UUID]time.Time{
		bob.ID: time.Now().Add(-30 * time.Second), // 直近に見かけた → オンライン
	}}
	uc := NewFriendUsecase(repo, presence)

	list, _, err := uc.List(ctx, alice.ID, FriendListFilter{Status: domain.FriendStatusAccepted, Limit: 10})
	if err != nil || len(list) != 1 {
		t.Fatalf("1 件返るはず: %v %d", err, len(list))
	}
	if !list[0].Online {
		t.Error("直近に見かけた相手はオンラインのはず")
	}

	// 窓を超えたらオフライン。
	presence.seen[bob.ID] = time.Now().Add(-PresenceOnlineWindow - time.Second)
	list, _, _ = uc.List(ctx, alice.ID, FriendListFilter{Status: domain.FriendStatusAccepted, Limit: 10})
	if list[0].Online {
		t.Error("しばらく見かけない相手はオフラインのはず")
	}
}
