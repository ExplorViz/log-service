package logs

import (
	"encoding/json"
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
	mux.HandleFunc("GET /v3/landscapes/{landscapeToken}/logs", h.getLandscapeLogs)
	mux.HandleFunc("GET /v3/landscapes/{landscapeToken}/entities/{telemetryKey}/logs", h.getEntityLogs)
}

func (h *Handler) getLandscapeLogs(w http.ResponseWriter, r *http.Request) {
	lt := r.PathValue("landscapeToken")
	if lt == "" {
		http.Error(w, "Missing or invalid landscape token in path parameter", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	params := LogSearchParams{}

	msgBody := query.Get("messageBody")
	if msgBody != "" {
		params.MessageBody = &msgBody
	}

	serviceName := query.Get("serviceName")
	if serviceName != "" {
		params.ServiceName = &serviceName
	}

	minSeverity, err := strconv.ParseUint(query.Get("minSeverity"), 10, 8)
	if err == nil {
		params.MinSeverity = &minSeverity
	}

	maxSeverity, err := strconv.ParseUint(query.Get("maxSeverity"), 10, 8)
	if err == nil {
		params.MaxSeverity = &maxSeverity
	}

	severityText := query.Get("severityText")
	if severityText != "" {
		params.SeverityText = &severityText
	}

	from, err := strconv.ParseUint(query.Get("from"), 10, 64)
	if err == nil {
		params.FromUnixNano = &from
	}

	to, err := strconv.ParseUint(query.Get("to"), 10, 64)
	if err == nil {
		params.ToUnixNano = &to
	}

	commit := query.Get("commit")
	if commit != "" {
		params.CommitHash = &commit
	}

	limit, err := strconv.ParseUint(query.Get("limit"), 10, 64)
	if err != nil {
		limit = 0
	}

	offset, err := strconv.ParseUint(query.Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	}

	logs, err := h.repo.findLogs(r.Context(), lt, params, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getEntityLogs(w http.ResponseWriter, r *http.Request) {
	lt := r.PathValue("landscapeToken")
	if lt == "" {
		http.Error(w, "Missing or invalid landscape token in path parameter", http.StatusBadRequest)
		return
	}

	telemetryKey := r.PathValue("telemetryKey")
	if telemetryKey == "" {
		http.Error(w, "Missing or invalid entity key in path parameter", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	params := LogSearchParams{}

	msgBody := query.Get("messageBody")
	if msgBody != "" {
		params.MessageBody = &msgBody
	}

	serviceName := query.Get("serviceName")
	if serviceName != "" {
		params.ServiceName = &serviceName
	}

	minSeverity, err := strconv.ParseUint(query.Get("minSeverity"), 10, 8)
	if err == nil {
		params.MinSeverity = &minSeverity
	}

	maxSeverity, err := strconv.ParseUint(query.Get("maxSeverity"), 10, 8)
	if err == nil {
		params.MinSeverity = &maxSeverity
	}

	severityText := query.Get("severityText")
	if severityText != "" {
		params.SeverityText = &severityText
	}

	from, err := strconv.ParseUint(query.Get("from"), 10, 64)
	if err == nil {
		params.FromUnixNano = &from
	}

	to, err := strconv.ParseUint(query.Get("to"), 10, 64)
	if err == nil {
		params.ToUnixNano = &to
	}

	commit := query.Get("commit")
	if commit != "" {
		params.CommitHash = &commit
	}

	limit, err := strconv.ParseUint(query.Get("limit"), 10, 64)
	if err != nil {
		limit = 0
	}

	offset, err := strconv.ParseUint(query.Get("offset"), 10, 64)
	if err != nil {
		offset = 0
	}

	logs, err := h.repo.findLogs(r.Context(), lt, params, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
