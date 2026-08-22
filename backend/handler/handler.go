// 外側の世界（Webブラウザやスマホ）と、内側の世界（ユースケース）の「翻訳家」の役割を果たします。
// 外側から受け取ったデータをユースケースが理解できる形に変換し、逆にユースケースから返ってきた結果を、外側の世界が理解できる形に変換します。
// 【概念の例：書籍の購入】
//
//	ユーザーのブラウザから「{"book_id": 123, "action": "buy"}」というJSONデータ（HTTPリクエスト）が送られてきます。
//	ハンドラーは、この通信データを受け取り、「なるほど、ID:123の本を買いたいのだな」と解釈して、ユースケースの「購入シナリオ」にデータを渡します。
//	ユースケースから「購入成功」という結果を受け取ったら、それを「HTTPステータス 200 OK」というWeb用の通信言語に翻訳してブラウザに返します。
//
// ビジネスルールは一切持ちませんが、Webや通信の知識を持っています。
package handler

import (
	"net/http"

	"github.com/google/uuid"

	"transcendence-backend/domain"
	"transcendence-backend/usecase"
)

type Config struct {
	Auth AuthConfig
	Game GameConfig
}

type Handlers struct {
	Ping         *PingHandler
	Auth         *AuthHandler
	Match        *MatchHandler
	Stats        *StatsHandler
	Achievements *AchievementHandler
	Game         *GameHandler
	User         *UserHandler
	Media        *MediaHandler
	Friend       *FriendHandler
	APIKey       *APIKeyHandler
	Public       *PublicHandler
}

// PresenceToucher は「このユーザーを見かけた」を記録する先。
type PresenceToucher interface {
	Touch(userID uuid.UUID)
}

func NewHandlers(services usecase.Services, cfg Config, events EventSubscriber, presence PresenceToucher) Handlers {
	auth := NewAuthHandler(services.Auth, cfg.Auth)

	// 認証は AuthHandler の実装をそのまま借りる。Cookie の読み方を 2 箇所に書かないため。
	// 認証が通るたびに presence を更新して、フレンドのオンライン表示の材料にする。
	currentUser := auth.currentUser
	if presence != nil {
		currentUser = func(r *http.Request) (*domain.User, error) {
			user, err := auth.currentUser(r)
			if err == nil {
				presence.Touch(user.ID)
			}
			return user, err
		}
	}

	return Handlers{
		Ping:         NewPingHandler(services.Ping),
		Auth:         auth,
		Match:        NewMatchHandler(services.Match, currentUser),
		Stats:        NewStatsHandler(services.Stats, currentUser),
		Achievements: NewAchievementHandler(services.Achievements, events, currentUser),
		// WebSocket のハンドシェイクも Cookie セッションで認証する。
		Game:   NewGameHandler(services.Game, currentUser, cfg.Game, presence),
		User:   NewUserHandler(services.User, currentUser),
		Media:  NewMediaHandler(services.Media, currentUser),
		Friend: NewFriendHandler(services.Friend, currentUser),
		APIKey: NewAPIKeyHandler(services.APIKey, currentUser),
		Public: NewPublicHandler(services.APIKey, services.User, services.Match, services.Stats, services.Friend),
	}
}

func NewRouter(handlers Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Ping.Ping)

	// Go 1.22 以降の ServeMux はメソッド付きパターンを解釈し、
	// より具体的なパターンを優先する。"/" と共存できる。
	mux.HandleFunc("GET /auth/google", handlers.Auth.Start)
	mux.HandleFunc("GET /auth/google/callback", handlers.Auth.Callback)
	mux.HandleFunc("GET /auth/me", handlers.Auth.Me)
	mux.HandleFunc("POST /auth/logout", handlers.Auth.Logout)

	// export.csv は "/matches/" 配下ではなく完全一致で登録する。
	// より具体的なパターンが優先されるので "GET /matches" と共存できる。
	mux.HandleFunc("POST /matches", handlers.Match.Create)
	mux.HandleFunc("GET /matches", handlers.Match.List)
	mux.HandleFunc("GET /matches/export.csv", handlers.Match.ExportCSV)

	mux.HandleFunc("GET /stats/me", handlers.Stats.Summary)
	mux.HandleFunc("GET /stats/me/timeseries", handlers.Stats.Timeseries)
	mux.HandleFunc("GET /stats/me/breakdown", handlers.Stats.Breakdown)
	mux.HandleFunc("GET /leaderboard", handlers.Stats.Leaderboard)

	mux.HandleFunc("GET /achievements/me", handlers.Achievements.List)
	mux.HandleFunc("GET /stats/stream", handlers.Achievements.Stream)

	mux.HandleFunc("GET /game/ws", handlers.Game.WS)
	mux.HandleFunc("GET /game/live", handlers.Game.Live)

	// /users/me はリテラル優先で {userId} と共存できる。
	mux.HandleFunc("GET /users/me", handlers.User.Me)
	mux.HandleFunc("PATCH /users/me", handlers.User.UpdateMe)
	mux.HandleFunc("GET /users", handlers.User.List)
	mux.HandleFunc("GET /users/{userId}", handlers.User.Get)

	mux.HandleFunc("POST /media/avatars", handlers.Media.UploadAvatar)
	mux.HandleFunc("GET /media", handlers.Media.List)
	mux.HandleFunc("GET /media/{assetId}", handlers.Media.Get)
	mux.HandleFunc("DELETE /media/{assetId}", handlers.Media.Delete)
	mux.HandleFunc("GET /media/{assetId}/file", handlers.Media.File)

	mux.HandleFunc("GET /friends", handlers.Friend.List)
	mux.HandleFunc("GET /friends/requests", handlers.Friend.ListRequests)
	mux.HandleFunc("POST /friends/requests", handlers.Friend.CreateRequest)
	mux.HandleFunc("PATCH /friends/requests/{userId}", handlers.Friend.Decide)
	mux.HandleFunc("DELETE /friends/{userId}", handlers.Friend.Remove)

	// API キー管理は Cookie セッションで守る。キー自身でキーは作れない。
	mux.HandleFunc("POST /apikeys", handlers.APIKey.Create)
	mux.HandleFunc("GET /apikeys", handlers.APIKey.List)
	mux.HandleFunc("DELETE /apikeys/{keyId}", handlers.APIKey.Revoke)

	// Public API（/v1）は API キーで認証する。
	handlers.Public.Register(mux)

	return mux
}
