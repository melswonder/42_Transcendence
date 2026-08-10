// 最も外側の層。usecase が宣言した interface を実装する。
package infrastructure

import (
	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

type PingRepo struct{}

// usecase.PingRepository を満たしているかコンパイル時に検査。
var _ usecase.PingRepository = (*PingRepo)(nil)

func NewPingRepo() *PingRepo {
	return &PingRepo{}
}

func (r *PingRepo) Get() domain.Ping {
	return domain.Ping{Message: "pong"}
}
