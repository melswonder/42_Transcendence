package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"transcendence-backend/usecase"
)

// PingHandler は疎通確認の HTTP 入口。
type PingHandler struct {
	uc *usecase.PingUsecase
}

func NewPingHandler(uc *usecase.PingUsecase) *PingHandler {
	return &PingHandler{uc: uc}
}

func (h *PingHandler) Ping(w http.ResponseWriter, _ *http.Request) {
	p := h.uc.Ping()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": p.Message}); err != nil {
		log.Printf("failed to encode ping response: %v", err)
	}
}
