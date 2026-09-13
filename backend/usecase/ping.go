package usecase

import "transcendence-backend/domain"

type PingRepository interface {
	Get() domain.Ping
}

type PingUsecase struct {
	repo PingRepository
}

func NewPingUsecase(repo PingRepository) *PingUsecase {
	return &PingUsecase{repo: repo}
}

func (u *PingUsecase) Ping() domain.Ping {
	return u.repo.Get()
}
