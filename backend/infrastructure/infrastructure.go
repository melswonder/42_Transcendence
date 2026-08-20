// 一番外側に位置し、「技術的な詳細」を担当する層です。
// データベース、外部のAPI、Webフレームワークなど、具体的なツールやハードウェアと直接やり取りをします。
// 【概念の例：書籍の購入】
//
//	データベース: ユースケースが「保存して」と言ったデータを、実際にPostgreSQLなどの具体的なデータベースに書き込むSQLを発行します。
//	外部システム: 決済の際に、実際にStripeやクレジットカード会社のAPIと通信を行います。
//
// システムのコア（ドメインやユースケース）から見れば、「裏側でどんなDBや決済ツールが動いているか」
// は隠蔽されており、インフラ層だけがその具体的な手段を知っています。
package infrastructure

import (
	"gorm.io/gorm"

	"transcendence-backend/usecase"
)

// Config は infrastructure 層が外の世界と話すために要る設定。
type Config struct {
	Google GoogleOAuthConfig
}

type Repositories struct {
	db          *gorm.DB
	Ping        usecase.PingRepository
	Auth        usecase.AuthRepository
	Match       usecase.MatchRepository
	Stats       usecase.StatsRepository
	GoogleOAuth usecase.OAuthProvider
}

func NewRepositories(db *gorm.DB, cfg Config) Repositories {
	return Repositories{
		db:          db,
		Ping:        NewPingRepo(),
		Auth:        NewAuthRepo(db),
		Match:       NewMatchRepo(db),
		Stats:       NewStatsRepo(db),
		GoogleOAuth: NewGoogleOAuth(cfg.Google),
	}
}

// Dependencies は各リポジトリを usecase 側の受け口へ詰め替える。
func (r Repositories) Dependencies() usecase.Dependencies {
	return usecase.Dependencies{
		PingRepository:  r.Ping,
		AuthRepository:  r.Auth,
		MatchRepository: r.Match,
		StatsRepository: r.Stats,
		GoogleOAuth:     r.GoogleOAuth,
	}
}
