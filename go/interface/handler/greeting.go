// HTTP と内部を橋渡しする層。JSON への変換だけ担い、処理は usecase に任せる。
package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"transcendence-backend/usecase"
)

type GreetingHandler struct {
	uc *usecase.GreetingUsecase
}

func NewGreetingHandler(uc *usecase.GreetingUsecase) *GreetingHandler {
	return &GreetingHandler{uc: uc}
}

func (h *GreetingHandler) Hello(w http.ResponseWriter, _ *http.Request) {
	g := h.uc.Hello()
	w.Header().Set("Content-Type", "application/json")
	// ここまででヘッダーは送信済みなのでステータスは変えられない。握りつぶさずログに残す。
	if err := json.NewEncoder(w).Encode(map[string]string{"message": g.Message}); err != nil {
		log.Printf("failed to encode greeting response: %v", err)
	}
}
