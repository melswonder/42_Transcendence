package main

import (
	"log"
	"net/http"
	"os"

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
func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		next(w, r)
	}
}

func main() {
	// --- 依存の配線（composition root）---
	// 外側の具象を生成し、内側へ interface として注入していく。
	repo := infrastructure.NewGreetingRepo() // infrastructure（具象）
	uc := usecase.NewGreetingUsecase(repo)   // usecase に repo を注入
	h := handler.NewGreetingHandler(uc)      // handler に usecase を注入

	// --- ルーティング ---
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.Hello)

	// CORS をアプリ全体に適用
	root := cors(mux.ServeHTTP)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, root); err != nil {
		log.Fatal(err)
	}
}
