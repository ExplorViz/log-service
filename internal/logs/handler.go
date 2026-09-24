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
}

func (h *Handler) getLandscapeLogs(w http.ResponseWriter, r *http.Request) {
	lt := r.PathValue("landscapeToken")
	if lt == "" {
		http.Error(w, "Missing or invalid landscape token in path parameter", http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	params := LogSearchParams{
		MessageBody:       strOrNil(query.Get("messageBody")),
		IncludeAttribKeys: query.Get("includeAttributeKeys") == "true",
		IncludeAttribVals: query.Get("includeAttributeValues") == "true",
		ServiceName:       strOrNil(query.Get("serviceName")),
		TelemetryKey:      strOrNil(query.Get("telemetryKey")),
		MinSeverity:       parseUintOrNil(query.Get("minSeverity")),
		MaxSeverity:       parseUintOrNil(query.Get("maxSeverity")),
		SeverityText:      strOrNil(query.Get("severityText")),
		TraceID:           strOrNil(query.Get("traceId")),
		SpanID:            strOrNil(query.Get("spanId")),
		FromUnixNano:      parseUintOrNil(query.Get("from")),
		ToUnixNano:        parseUintOrNil(query.Get("to")),
		CommitHash:        strOrNil(query.Get("commit")),
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

	cursorID := query.Get("cursorId")
	cursorTs := query.Get("cursorTimestamp")
	cursorSev := query.Get("cursorSeverity")

	if cursorID != "" && cursorTs != "" && cursorSev != "" {
		var parsedTs uint64
		if parsedTs, err = strconv.ParseUint(cursorTs, 10, 64); err != nil {
			http.Error(w, "Cursor timestamp is not valid Uint64", http.StatusBadRequest)
			return
		}

		var parsedSev uint64
		if parsedSev, err = strconv.ParseUint(cursorSev, 10, 64); err != nil {
			http.Error(w, "Cursor severity is not valid Uint64", http.StatusBadRequest)
			return
		}

		params.Cursor = &LogSearchCursor{
			LogID:          cursorID,
			Timestamp:      parsedTs,
			SeverityNumber: parsedSev,
		}
	} else if cursorID != "" || cursorTs != "" || cursorSev != "" {
		http.Error(w, "Provided some, but not all cursor values", http.StatusBadRequest)
		return
	}

	logs, err := h.repo.findLogs(r.Context(), lt, params, limit)
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

func strOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseUintOrNil(s string) *uint64 {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}
