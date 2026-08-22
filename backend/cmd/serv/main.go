package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

// corsMiddleware は許可したオリジンにだけ CORS ヘッダを付ける Gin ミドルウェア。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			// 資格情報（Cookie）付きでは、ブラウザは Allow-Methods / Allow-Headers の
			// ワイルドカード "*" を展開しない。preflight を通すため必ず明示する。
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			if h := c.GetHeader("Access-Control-Request-Headers"); h != "" {
				// preflight が尋ねてきたヘッダをそのまま許可して返す。
				c.Header("Access-Control-Allow-Headers", h)
			} else {
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			c.Header("Vary", "Origin")
		}

		// プリフライトリクエスト
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// accessLog は 1 リクエスト 1 行のアクセスログ。
// SSE / WebSocket のような張りっぱなしの接続もあるので、接続が閉じたときに出る。
func accessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Printf("%s %s -> %d (%s)",
			c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start).Round(time.Millisecond))
	}
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

// envIntOr は整数の環境変数。無効な値や未設定は fallback。
func envIntOr(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
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
			MediaDir: envOr("MEDIA_DIR", "./uploads"),
			// Public API の流量。デモしやすいよう環境変数で下げられる。
			PublicRateLimit: envIntOr("PUBLIC_API_RATE_LIMIT", 60),
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
	repositories, err := infrastructure.NewRepositories(db, cfg.infrastructure)
	if err != nil {
		log.Fatalf("failed to build repositories: %v", err)
	}
	services := usecase.NewServices(repositories.Dependencies())
	handlers := handler.NewHandlers(services, cfg.handler, repositories.Events, repositories.Presence)

	// recovery を最初に置き、パニックでもプロセスを落とさない。
	return handler.NewRouter(handlers, gin.Recovery(), accessLog(), corsMiddleware())
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
