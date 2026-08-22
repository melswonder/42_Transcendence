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

	"github.com/gin-gonic/gin"
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

// wrapF は net/http のハンドラを Gin のルートに載せるアダプタ。
// c.Param を r.PathValue に詰め替えるので、各ハンドラは Gin を知らずに済む。
func wrapF(h http.HandlerFunc, params ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, p := range params {
			c.Request.SetPathValue(p, c.Param(p))
		}
		h(c.Writer, c.Request)
	}
}

// NewRouter は全ルートを Gin で組み立てる。
// ルート木とミドルウェアは Gin、リクエストの解釈と応答は各ハンドラの担当。
func NewRouter(handlers Handlers, middleware ...gin.HandlerFunc) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware...)

	// 未マッチのパスは疎通確認の ping に落とす。
	router.NoRoute(wrapF(handlers.Ping.Ping))

	auth := router.Group("/auth")
	{
		auth.GET("/google", wrapF(handlers.Auth.Start))
		auth.GET("/google/callback", wrapF(handlers.Auth.Callback))
		auth.POST("/register", wrapF(handlers.Auth.Register))
		auth.POST("/login", wrapF(handlers.Auth.Login))
		auth.GET("/me", wrapF(handlers.Auth.Me))
		auth.POST("/logout", wrapF(handlers.Auth.Logout))
	}

	matches := router.Group("/matches")
	{
		matches.POST("", wrapF(handlers.Match.Create))
		matches.GET("", wrapF(handlers.Match.List))
		matches.GET("/export.csv", wrapF(handlers.Match.ExportCSV))
	}

	stats := router.Group("/stats")
	{
		stats.GET("/me", wrapF(handlers.Stats.Summary))
		stats.GET("/me/timeseries", wrapF(handlers.Stats.Timeseries))
		stats.GET("/me/breakdown", wrapF(handlers.Stats.Breakdown))
		stats.GET("/opponents", wrapF(handlers.Stats.Opponents))
		stats.GET("/stream", wrapF(handlers.Achievements.Stream))
	}
	router.GET("/leaderboard", wrapF(handlers.Stats.Leaderboard))
	router.GET("/achievements/me", wrapF(handlers.Achievements.List))

	game := router.Group("/game")
	{
		game.GET("/ws", wrapF(handlers.Game.WS))
		game.GET("/live", wrapF(handlers.Game.Live))
	}

	users := router.Group("/users")
	{
		users.GET("/me", wrapF(handlers.User.Me))
		users.PATCH("/me", wrapF(handlers.User.UpdateMe))
		users.GET("", wrapF(handlers.User.List))
		users.GET("/:userId", wrapF(handlers.User.Get, "userId"))
	}

	media := router.Group("/media")
	{
		media.POST("/avatars", wrapF(handlers.Media.UploadAvatar))
		media.GET("", wrapF(handlers.Media.List))
		media.GET("/:assetId", wrapF(handlers.Media.Get, "assetId"))
		media.DELETE("/:assetId", wrapF(handlers.Media.Delete, "assetId"))
		media.GET("/:assetId/file", wrapF(handlers.Media.File, "assetId"))
	}

	friends := router.Group("/friends")
	{
		friends.GET("", wrapF(handlers.Friend.List))
		friends.GET("/requests", wrapF(handlers.Friend.ListRequests))
		friends.POST("/requests", wrapF(handlers.Friend.CreateRequest))
		friends.PATCH("/requests/:userId", wrapF(handlers.Friend.Decide, "userId"))
		friends.DELETE("/:userId", wrapF(handlers.Friend.Remove, "userId"))
	}

	apikeys := router.Group("/apikeys")
	{
		apikeys.POST("", wrapF(handlers.APIKey.Create))
		apikeys.GET("", wrapF(handlers.APIKey.List))
		apikeys.DELETE("/:keyId", wrapF(handlers.APIKey.Revoke, "keyId"))
	}

	handlers.Public.Register(router)

	return router
}
