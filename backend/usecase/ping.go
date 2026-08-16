package usecase

import "transcendence-backend/domain"

// PingRepository
// Pingコマンドで使う関数
type PingRepository interface {
	Get() domain.Ping
}

// PingUsecase は疎通確認の流れを進める。
// Ping はリポジトリからメッセージを取って、そのまま返す。
type PingUsecase struct {
	repo PingRepository
}

func NewPingUsecase(repo PingRepository) *PingUsecase {
	return &PingUsecase{repo: repo}
}

func (u *PingUsecase) Ping() domain.Ping {
	return u.repo.Get()
}
