package instrumentation

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

// Starts and returns a new span representing a ClickHouse database operation, with attributes set according to the [database semantic conventions].
// Attributes are set for the database system, the query (using the provided value) and its parameters (using the provided value).
// The span must be ended manually. The queryParams slice should contain [driver.NamedValue] values, values of a different type are skipped.
//
// [database semantic conventions]: https://opentelemetry.io/docs/specs/semconv/db/database-spans/
func StartClickHouseSpan(ctx context.Context, name string, query string, queryParams []any) trace.Span {
	tracer := otel.Tracer("log-service")
	_, span := tracer.Start(
		ctx,
		name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemNameClickHouse,
			semconv.DBQueryText(query),
		),
	)
	for _, p := range queryParams {
		param, ok := p.(driver.NamedValue)
		if !ok {
			continue
		}
		span.SetAttributes(semconv.DBQueryParameter(param.Name, fmt.Sprint(param.Value)))
	}
	return span
}
