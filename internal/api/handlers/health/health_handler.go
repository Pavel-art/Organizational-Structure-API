package health

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/Pavel-art/Organizational-Structure-API/pkg/response"
)

const (
	healthTimeout = 2 * time.Second
)

type Handler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		response.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db down"})
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
