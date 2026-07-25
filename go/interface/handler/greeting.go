// HTTP と内部を橋渡しする層。JSON への変換だけ担い、処理は usecase に任せる。
package handler

import (
	"encoding/json"
	"net/http"

	"transcendence-backend/usecase"
)

type GreetingHandler struct {
	uc *usecase.GreetingUsecase
}

func NewGreetingHandler(uc *usecase.GreetingUsecase) *GreetingHandler {
	return &GreetingHandler{uc: uc}
}

func (h *GreetingHandler) Hello(w http.ResponseWriter, r *http.Request) {
	g := h.uc.Hello()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": g.Message})
}
