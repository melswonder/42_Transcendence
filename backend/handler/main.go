// 外側の世界（Webブラウザやスマホ）と、内側の世界（ユースケース）の「翻訳家」の役割を果たします。
// 外側から受け取ったデータをユースケースが理解できる形に変換し、逆にユースケースから返ってきた結果を、外側の世界が理解できる形に変換します。
// 【概念の例：書籍の購入】
//     ユーザーのブラウザから「{"book_id": 123, "action": "buy"}」というJSONデータ（HTTPリクエスト）が送られてきます。
//     ハンドラーは、この通信データを受け取り、「なるほど、ID:123の本を買いたいのだな」と解釈して、ユースケースの「購入シナリオ」にデータを渡します。
//     ユースケースから「購入成功」という結果を受け取ったら、それを「HTTPステータス 200 OK」というWeb用の通信言語に翻訳してブラウザに返します。
// ビジネスルールは一切持ちませんが、Webや通信の知識を持っています。
package handler

import (
	"net/http"

	"transcendence-backend/usecase"
)

type Handlers struct {
	Ping *PingHandler
}

func NewHandlers(services usecase.Services) Handlers {
	return Handlers{
		Ping: NewPingHandler(services.Ping),
	}
}

func NewRouter(handlers Handlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Ping.Ping)

	return mux
}
