package logs

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

type Handler struct {
	repo Repository
}

func NewHandler(r Repository) Handler {
	return Handler{
		repo: r,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /v3/landscapes/{landscapeToken}/entities/{telemetryKey}/logs", h.getEntityLogs)
}

func (h *Handler) getEntityLogs(w http.ResponseWriter, r *http.Request) {
	lt := r.PathValue("landscapeToken")
	if lt == "" {
		http.Error(w, "Missing or invalid landscape token in path parameter", http.StatusBadRequest)
		return
	}

	telemetryKey := r.PathValue("telemetryKey")
	if telemetryKey == "" {
		http.Error(w, "Missing or invalid telemetry key in path parameter", http.StatusBadRequest)
		return
	}

	from, err := strconv.ParseUint(r.URL.Query().Get("from"), 10, 64)
	if err != nil {
		from = 0
	}

	to, err := strconv.ParseUint(r.URL.Query().Get("to"), 10, 64)
	if err != nil {
		to = math.MaxUint64
	}

	commit := r.URL.Query().Get("commit")

	logs, err := h.repo.findEntityLogs(r.Context(), lt, telemetryKey, from, to, commit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
