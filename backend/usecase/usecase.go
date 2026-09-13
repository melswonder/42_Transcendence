// Package usecase はシナリオを進める層。domain のルールを呼び、永続化は
// interface 越しに依頼する（どの DB に保存されるかは知らない）。
package usecase

type Dependencies struct {
	PingRepository        PingRepository
	AuthRepository        AuthRepository
	MatchRepository       MatchRepository
	StatsRepository       StatsRepository
	AchievementRepository AchievementRepository
	GameRepository        GameRepository
	UserRepository        UserRepository
	MediaRepository       MediaRepository
	MediaFileStore        FileStore
	FriendRepository      FriendRepository
	Presence              PresenceTracker
	APIKeyRepository      APIKeyRepository
	RateLimiter           RateLimiter
	GoogleOAuth           OAuthProvider
	MatchNotifier         MatchNotifier
}

type Services struct {
	Ping         *PingUsecase
	Auth         *AuthUsecase
	Match        *MatchUsecase
	Stats        *StatsUsecase
	Achievements *AchievementUsecase
	Game         *GameUsecase
	User         *UserUsecase
	Media        *MediaUsecase
	Friend       *FriendUsecase
	APIKey       *APIKeyUsecase
}

func NewServices(dependencies Dependencies) Services {
	// 実績の判定は統計の集計に乗るので、先に Stats を組んでから渡す。
	stats := NewStatsUsecase(dependencies.StatsRepository)
	achievements := NewAchievementUsecase(dependencies.AchievementRepository, stats)

	return Services{
		Ping:         NewPingUsecase(dependencies.PingRepository),
		Auth:         NewAuthUsecase(dependencies.GoogleOAuth, dependencies.AuthRepository),
		Match:        NewMatchUsecase(dependencies.MatchRepository, dependencies.MatchNotifier, achievements),
		Stats:        stats,
		Achievements: achievements,
		Game:         NewGameUsecase(dependencies.GameRepository, dependencies.MatchNotifier, achievements),
		User:         NewUserUsecase(dependencies.UserRepository),
		Media:        NewMediaUsecase(dependencies.MediaRepository, dependencies.MediaFileStore),
		Friend:       NewFriendUsecase(dependencies.FriendRepository, dependencies.Presence),
		APIKey:       NewAPIKeyUsecase(dependencies.APIKeyRepository, dependencies.RateLimiter),
	}
}
