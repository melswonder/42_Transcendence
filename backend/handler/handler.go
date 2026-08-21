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

	"transcendence-backend/usecase"
)

type Config struct {
	Auth AuthConfig
	Game GameConfig
}

type Handlers struct {
	Ping *PingHandler
	Auth *AuthHandler
	Game *GameHandler
}

func NewHandlers(services usecase.Services, cfg Config) Handlers {
	auth := NewAuthHandler(services.Auth, cfg.Auth)
	return Handlers{
		Ping: NewPingHandler(services.Ping),
		Auth: auth,
		// WebSocket のハンドシェイクも Cookie セッションで認証する。
		Game: NewGameHandler(services.Game, auth.currentUser, cfg.Game),
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

	mux.HandleFunc("GET /game/ws", handlers.Game.WS)

	return mux
}
