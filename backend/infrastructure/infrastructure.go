// 最も外側の層。usecase が宣言した interface を DB・外部 API・ファイル保存で実装する。
package infrastructure

import (
	"gorm.io/gorm"

	"transcendence-backend/usecase"
)

// Config は infrastructure 層が外の世界と話すために要る設定。
type Config struct {
	Google GoogleOAuthConfig
	// MediaDir はアップロードされたファイルの保存先ディレクトリ。
	MediaDir string
	// PublicRateLimit は Public API のキーごとの流量（リクエスト/分）。
	PublicRateLimit int
}

type Repositories struct {
	db           *gorm.DB
	Ping         usecase.PingRepository
	Auth         usecase.AuthRepository
	Match        usecase.MatchRepository
	Stats        usecase.StatsRepository
	Achievements usecase.AchievementRepository
	Game         usecase.GameRepository
	User         usecase.UserRepository
	Media        usecase.MediaRepository
	MediaFiles   usecase.FileStore
	Friend       usecase.FriendRepository
	APIKeys      usecase.APIKeyRepository
	RateLimit    usecase.RateLimiter
	GoogleOAuth  usecase.OAuthProvider
	// Presence は「最後に見かけた時刻」のメモリ上の記録。オンライン表示に使う。
	Presence *PresenceHub
	// Events は SSE の配信元。リポジトリではないが、組み立てる場所が同じなのでここで持つ。
	Events *EventHub
}

func NewRepositories(db *gorm.DB, cfg Config) (Repositories, error) {
	files, err := NewLocalFileStore(cfg.MediaDir)
	if err != nil {
		return Repositories{}, err
	}
	return Repositories{
		db:           db,
		Ping:         NewPingRepo(),
		Auth:         NewAuthRepo(db),
		Match:        NewMatchRepo(db),
		Stats:        NewStatsRepo(db),
		Achievements: NewAchievementRepo(db),
		Game:         NewGameRepo(db),
		User:         NewUserRepo(db),
		Media:        NewMediaRepo(db),
		MediaFiles:   files,
		Friend:       NewFriendRepo(db),
		APIKeys:      NewAPIKeyRepo(db),
		RateLimit:    NewFixedWindowLimiter(cfg.PublicRateLimit),
		GoogleOAuth:  NewGoogleOAuth(cfg.Google),
		Events:       NewEventHub(),
		Presence:     NewPresenceHub(),
	}, nil
}

// Dependencies は各リポジトリを usecase 側の受け口へ詰め替える。
func (r Repositories) Dependencies() usecase.Dependencies {
	return usecase.Dependencies{
		PingRepository:        r.Ping,
		AuthRepository:        r.Auth,
		MatchRepository:       r.Match,
		StatsRepository:       r.Stats,
		AchievementRepository: r.Achievements,
		GameRepository:        r.Game,
		UserRepository:        r.User,
		MediaRepository:       r.Media,
		MediaFileStore:        r.MediaFiles,
		FriendRepository:      r.Friend,
		Presence:              r.Presence,
		APIKeyRepository:      r.APIKeys,
		RateLimiter:           r.RateLimit,
		GoogleOAuth:           r.GoogleOAuth,
		MatchNotifier:         r.Events,
	}
}
