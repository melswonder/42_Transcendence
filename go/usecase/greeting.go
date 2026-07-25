// 操作の流れを置く層。domain には依存するが、DB や HTTP は知らない。
package usecase

import "transcendence-backend/domain"

// GreetingRepository は「欲しい機能」を interface で宣言する（繋ぎ目）。
// 実装は infrastructure が提供する。
type GreetingRepository interface {
	Get() domain.Greeting
}

type GreetingUsecase struct {
	repo GreetingRepository
}

func NewGreetingUsecase(repo GreetingRepository) *GreetingUsecase {
	return &GreetingUsecase{repo: repo}
}

func (u *GreetingUsecase) Hello() domain.Greeting {
	return u.repo.Get()
}
