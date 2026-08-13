package main

import (
	"log"
	"net/http"
	"os"

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

// newApplicationHandler は composition root。
// 具体的な infrastructure の実装を選び、内側の層へ繋ぐのはここだけ。
func newApplicationHandler(db *gorm.DB) http.Handler {
	repositories := infrastructure.NewRepositories(db)
	services := usecase.NewServices(repositories.Dependencies())
	handlers := handler.NewHandlers(services)

	return cors(handler.NewRouter(handlers))
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := infrastructure.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	log.Print("database connected")

	root := newApplicationHandler(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, root); err != nil {
		log.Fatal(err)
	}
}
