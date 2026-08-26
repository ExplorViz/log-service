package logs

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Repository struct {
	Conn driver.Conn
}

type LogSearchParams struct {
	MessageBody       *string
	IncludeAttribKeys bool
	IncludeAttribVals bool

	TelemetryKey *string
	ServiceName  *string

	MinSeverity  *uint64
	MaxSeverity  *uint64
	SeverityText *string

	FromUnixNano *uint64
	ToUnixNano   *uint64
	CommitHash   *string
	TraceID      *string
	SpanID       *string

	SortBy LogSearchSorting

	// Limits the number of retrieved rows, for use with paginiation.
	Limit *uint64

	// Specifies the last received log from the previous request.
	Cursor *LogSearchCursor
}

type LogSearchSorting int

const (
	SortNewest LogSearchSorting = iota
	SortOldest
	SortHighestSeverity
)

// A LogSearchCursor specifies the last seen log from a prior request.
// This can be used for pagination.
type LogSearchCursor struct {
	LogID          string
	Timestamp      uint64
	SeverityNumber uint64
}

// findLogs searches the database for logs associated with the given landscape. The search space can be restricted using
// a variety of filter options (see [LogSearchParams]). A limit can be specified for pagination.
func (r *Repository) findLogs(ctx context.Context, landscapeToken string, params LogSearchParams, limit uint64) ([]Log, error) {
	queryParams := make([]any, 0, 15)
	var conditions strings.Builder

	queryParams = append(queryParams, clickhouse.Named("landscapeToken", landscapeToken))

	if params.MessageBody != nil {
		conditions.WriteString(" AND (hasAllTokens(Body, @messageBody)")
		queryParams = append(queryParams, clickhouse.Named("messageBody", *params.MessageBody))

		if params.IncludeAttribKeys {
			conditions.WriteString(`
				OR (
					hasAllTokens(mapKeys(LogAttributes), @messageBody)
					OR hasAllTokens(mapKeys(ScopeAttributes), @messageBody)
					OR hasAllTokens(mapKeys(ResourceAttributes), @messageBody)
				)`)
		}

		if params.IncludeAttribVals {
			conditions.WriteString(`
				OR (
					hasAllTokens(mapValues(LogAttributes), @messageBody)
					OR hasAllTokens(mapValues(ScopeAttributes), @messageBody)
					OR hasAllTokens(mapValues(ResourceAttributes), @messageBody)
				)`)
		}
		conditions.WriteString(")")
	}

	if params.TelemetryKey != nil {
		conditions.WriteString(" AND ExplorvizTelemetryKey = @telemetryKey")
		queryParams = append(queryParams, clickhouse.Named("telemetryKey", *params.TelemetryKey))
	}

	if params.ServiceName != nil {
		conditions.WriteString(" AND ServiceName = @serviceName")
		queryParams = append(queryParams, clickhouse.Named("serviceName", *params.ServiceName))
	}

	if params.MinSeverity != nil {
		conditions.WriteString(" AND SeverityNumber >= @minSeverity")
		queryParams = append(queryParams, clickhouse.Named("minSeverity", *params.MinSeverity))
	}

	if params.MaxSeverity != nil {
		conditions.WriteString(" AND SeverityNumber <= @maxSeverity")
		queryParams = append(queryParams, clickhouse.Named("maxSeverity", *params.MaxSeverity))
	}

	if params.SeverityText != nil {
		conditions.WriteString(" AND SeverityText = @severityText")
		queryParams = append(queryParams, clickhouse.Named("severityText", *params.SeverityText))
	}

	if params.FromUnixNano != nil {
		conditions.WriteString(" AND Timestamp_ns >= @from")
		queryParams = append(queryParams, clickhouse.Named("from", *params.FromUnixNano))
	}

	if params.ToUnixNano != nil {
		conditions.WriteString(" AND Timestamp_ns < @to")
		queryParams = append(queryParams, clickhouse.Named("to", *params.ToUnixNano))
	}

	if params.CommitHash != nil {
		conditions.WriteString(" AND CommitHash = @commitHash")
		queryParams = append(queryParams, clickhouse.Named("commitHash", *params.CommitHash))
	}

	if params.TraceID != nil {
		conditions.WriteString(" AND TraceId = @traceId")
		queryParams = append(queryParams, clickhouse.Named("traceId", *params.TraceID))
	}

	if params.SpanID != nil {
		conditions.WriteString(" AND SpanId = @spanId")
		queryParams = append(queryParams, clickhouse.Named("spanId", *params.SpanID))
	}

	if params.Cursor != nil {
		switch params.SortBy {
		case SortNewest:
			conditions.WriteString(`
				AND (
					Timestamp_ns < @cursorTimestamp
					OR (Timestamp_ns = @cursorTimestamp AND LogId > @cursorId)
				)`)
		case SortOldest:
			conditions.WriteString(`
				AND (
					Timestamp_ns > @cursorTimestamp
					OR (Timestamp_ns = @cursorTimestamp AND LogId > @cursorId)
				)`)

		case SortHighestSeverity:
			conditions.WriteString(`
				AND (
					SeverityNumber < @cursorSeverity
					OR (
						SeverityNumber = @cursorSeverity
						AND (
							Timestamp_ns < @cursorTimestamp
							OR (Timestamp_ns = @cursorTimestamp AND LogId > @cursorId)
						)
					)
				)`)
		default:
			slog.Error("Received invalid sorting order", "SortBy", params.SortBy)
		}

		queryParams = append(queryParams,
			clickhouse.Named("cursorId", params.Cursor.LogID),
			clickhouse.Named("cursorTimestamp", params.Cursor.Timestamp),
			clickhouse.Named("cursorSeverity", params.Cursor.SeverityNumber),
		)
	}

	ordering := ""
	switch params.SortBy {
	case SortNewest:
		ordering += " ORDER BY Timestamp_ns DESC, LogId ASC"
	case SortOldest:
		ordering += " ORDER BY Timestamp_ns ASC, LogId ASC"
	case SortHighestSeverity:
		ordering += " ORDER BY SeverityNumber DESC, Timestamp_ns DESC, LogId ASC"
	default:
		slog.Error("Received invalid sorting order", "SortBy", params.SortBy)
	}

	queryLimit := ""
	if limit > 0 {
		queryLimit = " LIMIT @limit"
		queryParams = append(queryParams, clickhouse.Named("limit", limit))
	}

	logs := []Log{}

	err := r.Conn.Select(ctx, &logs, `
		SELECT
			toString(LogId) AS ID,
			Body AS MessageBody,
			ExplorvizTelemetryKey AS TelemetryKey,
			ServiceName,
			SeverityNumber AS Severity,
			SeverityText,
			Timestamp_ns AS TimeUnixNano,
			TraceId AS TraceID,
			SpanId AS SpanID,
			EventName,
			LogAttributes AS LogAttribs,
			ResourceAttributes AS ResourceAttribs
		FROM otel_logs
		WHERE
			ExplorvizTokenId = @landscapeToken
			`+conditions.String()+ordering+queryLimit, queryParams...)
	if err != nil {
		return []Log{}, err
	}

	return logs, nil
}

// findLogLevels searches the database for every distinct log level text that occurs within a given landscape.
func (r *Repository) findLogLevels(ctx context.Context, landscapeToken string) ([]string, error) {
	rows, err := r.Conn.Query(ctx, `
		SELECT DISTINCT SeverityText
		FROM otel_logs
		WHERE ExplorvizTokenId = @landscapeToken
		`, clickhouse.Named("landscapeToken", landscapeToken))
	if err != nil {
		return []string{}, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			fmt.Println("failed to close rows object")
		}
	}()

	severities := []string{}

	for rows.Next() {
		var severityName string
		if err := rows.Scan(&severityName); err != nil {
			return []string{}, err
		}

		if severityName != "" {
			severities = append(severities, severityName)
		}
	}

	return severities, nil
}
