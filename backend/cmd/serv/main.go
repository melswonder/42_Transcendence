package main

import (
	"log"
	"net/http"
	"os"

	"transcendence-backend/handler"
	"transcendence-backend/infrastructure"
	"transcendence-backend/usecase"
)

// APIの構築
func newApplicationHandler() http.Handler {

	repositories := infrastructure.NewRepositories()
	services := usecase.NewServices(repositories.Dependencies())
	handlers := handler.NewHandlers(services)

	return handler.NewRouter(handlers)
}

func main() {
	root := newApplicationHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Printf("server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, root); err != nil {
		log.Fatal(err)
	}
}
