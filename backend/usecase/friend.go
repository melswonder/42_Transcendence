package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"transcendence-backend/domain"
)

// PresenceOnlineWindow はこの時間内に活動があれば「オンライン」とみなす。
// HTTP は毎リクエスト、対局中は WebSocket の keep-alive（25 秒間隔）が活動になる。
const PresenceOnlineWindow = 2 * time.Minute

// PresenceTracker の実装はメモリ上のマップ。再起動で消えるが、presence は揮発してよい情報。
type PresenceTracker interface {
	Touch(userID uuid.UUID)
	LastSeen(userID uuid.UUID) (time.Time, bool)
}

// FriendPair は正規化を隠した 1 組の関係。
type FriendPair struct {
	Status        string
	RequestedByMe bool
}

type FriendListFilter struct {
	Status string
	// Direction は pending のときだけ意味を持つ。incoming = 相手からの申請。
	Direction string
	Limit     int
	Offset    int
}

// FriendRepository の実装は low/high の正規化を担う。
type FriendRepository interface {
	// GetPair は無ければ domain.ErrFriendshipNotFound。
	GetPair(ctx context.Context, me, other uuid.UUID) (*FriendPair, error)
	// InsertPending は既にあれば domain.ErrFriendAlreadyRequested。
	InsertPending(ctx context.Context, from, to uuid.UUID) error
	// UpdateStatus は requestedBy も同時に差し替える。
	UpdateStatus(ctx context.Context, me, other uuid.UUID, status string, requestedBy uuid.UUID) error
	// DeletePair は無ければ domain.ErrFriendshipNotFound。
	DeletePair(ctx context.Context, me, other uuid.UUID) error
	GetFriendship(ctx context.Context, me, other uuid.UUID) (*domain.Friendship, error)
	ListFriendships(ctx context.Context, me uuid.UUID, f FriendListFilter) ([]domain.Friendship, int, error)
	// HasBlockRelation はどちらか一方でもブロックしていれば true。
	HasBlockRelation(ctx context.Context, a, b uuid.UUID) (bool, error)
	// UserExists は実在するアクティブなユーザーなら true。
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
}

type FriendUsecase struct {
	repo     FriendRepository
	presence PresenceTracker
	now      func() time.Time
}

func NewFriendUsecase(repo FriendRepository, presence PresenceTracker) *FriendUsecase {
	return &FriendUsecase{repo: repo, presence: presence, now: time.Now}
}

func (u *FriendUsecase) List(ctx context.Context, me uuid.UUID, f FriendListFilter) ([]domain.Friendship, int, error) {
	friendships, total, err := u.repo.ListFriendships(ctx, me, f)
	if err != nil {
		return nil, 0, err
	}
	for i := range friendships {
		friendships[i].Online = u.isOnline(friendships[i].Other.ID)
	}
	return friendships, total, nil
}

// Request は相手が既に自分へ申請していれば、その場で accepted にする（相互申請の自動承認）。
func (u *FriendUsecase) Request(ctx context.Context, me, other uuid.UUID) (*domain.Friendship, error) {
	if me == other {
		return nil, domain.ErrFriendSelf
	}

	exists, err := u.repo.UserExists(ctx, other)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrUserNotFound
	}

	blocked, err := u.repo.HasBlockRelation(ctx, me, other)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, domain.ErrFriendBlocked
	}

	pair, err := u.repo.GetPair(ctx, me, other)
	switch {
	case err == nil:
		switch {
		case pair.Status == domain.FriendStatusAccepted:
			return nil, domain.ErrAlreadyFriends
		case pair.Status == domain.FriendStatusPending && pair.RequestedByMe:
			return nil, domain.ErrFriendAlreadyRequested
		case pair.Status == domain.FriendStatusPending:
			if err := u.repo.UpdateStatus(ctx, me, other, domain.FriendStatusAccepted, other); err != nil {
				return nil, err
			}
		default:
			// rejected からの再申請は申請者を差し替えて pending に戻す。
			if err := u.repo.UpdateStatus(ctx, me, other, domain.FriendStatusPending, me); err != nil {
				return nil, err
			}
		}
	case errors.Is(err, domain.ErrFriendshipNotFound):
		if err := u.repo.InsertPending(ctx, me, other); err != nil {
			return nil, err
		}
	default:
		return nil, err
	}

	return u.getOne(ctx, me, other)
}

func (u *FriendUsecase) Decide(ctx context.Context, me, other uuid.UUID, accept bool) (*domain.Friendship, error) {
	pair, err := u.repo.GetPair(ctx, me, other)
	if err != nil {
		return nil, err
	}
	// 自分宛の pending だけが対象。自分が送った申請は承認できない。
	if pair.Status != domain.FriendStatusPending || pair.RequestedByMe {
		return nil, domain.ErrFriendshipNotFound
	}

	status := domain.FriendStatusAccepted
	if !accept {
		status = domain.FriendStatusRejected
	}
	if err := u.repo.UpdateStatus(ctx, me, other, status, other); err != nil {
		return nil, err
	}
	return u.getOne(ctx, me, other)
}

// Remove は pending なら申請の取り下げにもなる。
func (u *FriendUsecase) Remove(ctx context.Context, me, other uuid.UUID) error {
	return u.repo.DeletePair(ctx, me, other)
}

func (u *FriendUsecase) isOnline(userID uuid.UUID) bool {
	if u.presence == nil {
		return false
	}
	seen, ok := u.presence.LastSeen(userID)
	return ok && u.now().Sub(seen) < PresenceOnlineWindow
}

// getOne は相手のユーザー情報を含めるため、操作後に引き直す。
func (u *FriendUsecase) getOne(ctx context.Context, me, other uuid.UUID) (*domain.Friendship, error) {
	friendship, err := u.repo.GetFriendship(ctx, me, other)
	if err != nil {
		return nil, err
	}
	friendship.Online = u.isOnline(other)
	return friendship, nil
}
