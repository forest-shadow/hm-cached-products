package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/forest-shadow/hm-cached-products/internal/logger"
)

type Handler struct {
	logger *logger.Logger
}

func New(logger *logger.Logger) http.Handler {
	handler := &Handler{logger: logger}
	return handler.routes()
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *Handler) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
