package logs

import (
	"encoding/json"
	"fmt"
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
	mux.HandleFunc("GET /v3/landscapes/{landscapeToken}/log-levels", h.getLandscapeLogLevels)
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

	if msgBody := query.Get("messageBody"); msgBody != "" {
		params.MessageBody = &msgBody
	}

	params.IncludeAttribKeys = query.Get("includeAttributeKeys") != ""
	params.IncludeAttribVals = query.Get("includeAttributeValues") != ""

	if serviceName := query.Get("serviceName"); serviceName != "" {
		params.ServiceName = &serviceName
	}

	if minSeverity, err := strconv.ParseUint(query.Get("minSeverity"), 10, 8); err == nil {
		params.MinSeverity = &minSeverity
	}

	if maxSeverity, err := strconv.ParseUint(query.Get("maxSeverity"), 10, 8); err == nil {
		params.MaxSeverity = &maxSeverity
	}

	if severityText := query.Get("severityText"); severityText != "" {
		params.SeverityText = &severityText
	}

	if traceID := query.Get("traceId"); traceID != "" {
		params.TraceID = &traceID
	}

	if spanID := query.Get("spanId"); spanID != "" {
		params.SpanID = &spanID
	}

	if from, err := strconv.ParseUint(query.Get("from"), 10, 64); err == nil {
		params.FromUnixNano = &from
	}

	if to, err := strconv.ParseUint(query.Get("to"), 10, 64); err == nil {
		params.ToUnixNano = &to
	}

	if commit := query.Get("commit"); commit != "" {
		params.CommitHash = &commit
	}

	switch sortBy := query.Get("sortBy"); sortBy {
	case "", "newest":
		params.SortBy = SortNewest
	case "oldest":
		params.SortBy = SortOldest
	case "severity":
		params.SortBy = SortHighestSeverity
	default:
		http.Error(w, fmt.Sprintf(`Invalid value %s for parameter "sortBy"`, sortBy), http.StatusBadRequest)
		return
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

func (h *Handler) getLandscapeLogLevels(w http.ResponseWriter, r *http.Request) {
	lt := r.PathValue("landscapeToken")
	if lt == "" {
		http.Error(w, "Missing or invalid landscape token in path parameter", http.StatusBadRequest)
		return
	}

	levels, err := h.repo.findLogLevels(r.Context(), lt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(levels); err != nil {
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

	if msgBody := query.Get("messageBody"); msgBody != "" {
		params.MessageBody = &msgBody
	}

	params.IncludeAttribKeys = query.Get("includeAttributeKeys") != ""
	params.IncludeAttribVals = query.Get("includeAttributeValues") != ""

	if serviceName := query.Get("serviceName"); serviceName != "" {
		params.ServiceName = &serviceName
	}

	if minSeverity, err := strconv.ParseUint(query.Get("minSeverity"), 10, 8); err == nil {
		params.MinSeverity = &minSeverity
	}

	if maxSeverity, err := strconv.ParseUint(query.Get("maxSeverity"), 10, 8); err == nil {
		params.MinSeverity = &maxSeverity
	}

	if severityText := query.Get("severityText"); severityText != "" {
		params.SeverityText = &severityText
	}

	if from, err := strconv.ParseUint(query.Get("from"), 10, 64); err == nil {
		params.FromUnixNano = &from
	}

	if to, err := strconv.ParseUint(query.Get("to"), 10, 64); err == nil {
		params.ToUnixNano = &to
	}

	if traceID := query.Get("traceId"); traceID != "" {
		params.TraceID = &traceID
	}

	if spanID := query.Get("spanId"); spanID != "" {
		params.SpanID = &spanID
	}

	if commit := query.Get("commit"); commit != "" {
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
