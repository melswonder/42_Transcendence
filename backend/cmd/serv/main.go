package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"gorm.io/gorm"

	"transcendence-backend/handler"
	"transcendence-backend/infrastructure"
	"transcendence-backend/usecase"
)

// CORSを許可するオリジン
var allowedOrigins = map[string]bool{
	"http://localhost:3000": true, // ローカル開発環境
	"http://frontend:3000":  true, // Dockerコンテナ間通信
}

// WebSocket のハンドシェイクを許可する Origin のホスト。
// WebSocket は CORS の対象外なので、上とは別に upgrade 時へ渡して検証する。
var allowedWSOrigins = []string{
	"localhost:3000",
	"frontend:3000",
}

// CORSミドルウェア
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")
		}

		// プリフライトリクエスト
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 必要な環境変数を入れる
type config struct {
	infrastructure infrastructure.Config
	handler        handler.Config
	port           string
}

// 環境変数から値を取り込むヘルパー関数
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is not set", key)
	}

	return v
}

// 環境変数が存在しない場合は fallbackを入れる
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// loadConfig は環境変数を読む、値が欠けていた場合にエラーにすることができる
// 　リクエストが来てエラーになるより、サーバーを構築する前にエラーにしたほうが安全
func loadConfig() config {
	frontendURL := envOr("FRONTEND_URL", "http://localhost:3000")

	return config{
		infrastructure: infrastructure.Config{
			Google: infrastructure.GoogleOAuthConfig{
				ClientID:     mustEnv("GOOGLE_CLIENT_ID"),
				ClientSecret: mustEnv("GOOGLE_CLIENT_SECRET"),
				RedirectURL:  mustEnv("GOOGLE_REDIRECT_URL"),
			},
		},
		handler: handler.Config{
			Auth: handler.AuthConfig{
				FrontendURL: frontendURL,
				// http:// のローカル開発では Secure を付けると Cookie が保存されない。
				SecureCookie: strings.HasPrefix(frontendURL, "https://"),
			},
			Game: handler.GameConfig{
				AllowedOrigins: allowedWSOrigins,
			},
		},
		port: envOr("PORT", "4000"),
	}
}

// dbの接続
func mustConnectDB() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := infrastructure.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	log.Print("database connected")

	return db
}

// newApplicationHandler は composition root。
// 具体的な infrastructure の実装を選び、内側の層へ繋ぐのはここだけ。
func newApplicationHandler(db *gorm.DB, cfg config) http.Handler {
	repositories := infrastructure.NewRepositories(db, cfg.infrastructure)
	services := usecase.NewServices(repositories.Dependencies())
	handlers := handler.NewHandlers(services, cfg.handler, repositories.Events)

	return cors(handler.NewRouter(handlers))
}

func main() {
	cfg := loadConfig()
	db := mustConnectDB()
	root := newApplicationHandler(db, cfg)

	log.Printf("server listening on :%s", cfg.port)
	if err := http.ListenAndServe(":"+cfg.port, root); err != nil {
		log.Fatal(err)
	}
}
