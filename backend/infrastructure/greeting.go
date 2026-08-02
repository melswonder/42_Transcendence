// 最も外側の層。usecase が宣言した interface を実装する。
package infrastructure

import (
	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

type GreetingRepo struct{}

// usecase.GreetingRepository を満たしているかコンパイル時に検査。
var _ usecase.GreetingRepository = (*GreetingRepo)(nil)

func NewGreetingRepo() *GreetingRepo {
	return &GreetingRepo{}
}

func (r *GreetingRepo) Get() domain.Greeting {
	return domain.Greeting{Message: "Hello from Go!"}
}
