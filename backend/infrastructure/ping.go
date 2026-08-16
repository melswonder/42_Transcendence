// 最も外側の層。usecase が宣言した interface を実装する。
package infrastructure

import (
	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

// PingRepo は usecase.PingRepository の実装。DB も外部サービスも使わない。
// Get は固定のメッセージ ("pong") を返す。
type PingRepo struct{}

// usecase.PingRepository を満たしているかコンパイル時に検査。
var _ usecase.PingRepository = (*PingRepo)(nil)

func NewPingRepo() *PingRepo {
	return &PingRepo{}
}

func (r *PingRepo) Get() domain.Ping {
	return domain.Ping{Message: "pong"}
}
