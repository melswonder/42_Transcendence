package domain

import (
	"errors"
	"time"
)

// friendships.status の値。
const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusRejected = "rejected"
)

// フレンドまわりの失敗。handler は errors.Is で 400 / 403 / 404 / 409 に振り分ける。
var (
	ErrFriendshipNotFound     = errors.New("friendship not found")
	ErrFriendSelf             = errors.New("cannot befriend yourself")
	ErrFriendBlocked          = errors.New("blocked relation exists")
	ErrFriendAlreadyRequested = errors.New("friend request already sent")
	ErrAlreadyFriends         = errors.New("already friends")
)

// Friendship は自分から見た 1 件のフレンド関係。
// DB の low/high 正規化は infrastructure に閉じ、この層では常に「相手」で表す。
type Friendship struct {
	Other         User
	Status        string
	RequestedByMe bool
	Online        bool // いまオンラインか（presence から導出）
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
