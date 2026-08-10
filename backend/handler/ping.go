// HTTP と内部を橋渡しする層。JSON への変換だけ担い、処理は usecase に任せる。
package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"transcendence-backend/usecase"
)

type PingHandler struct {
	uc *usecase.PingUsecase
}

func NewPingHandler(uc *usecase.PingUsecase) *PingHandler {
	return &PingHandler{uc: uc}
}

func (h *PingHandler) Ping(w http.ResponseWriter, _ *http.Request) {
	p := h.uc.Ping()
	w.Header().Set("Content-Type", "application/json")
	// ここまででヘッダーは送信済みなのでステータスは変えられない。握りつぶさずログに残す。
	if err := json.NewEncoder(w).Encode(map[string]string{"message": p.Message}); err != nil {
		log.Printf("failed to encode ping response: %v", err)
	}
}
