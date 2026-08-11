package logs

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Repository struct {
	Conn driver.Conn
}

// findEntityLogs searches the database for any logs belonging to the entity with the given telemetry key that occur
// within the time span given by fromUnixNano (inclusive) and toUnixNano (exclusive). To restrict search for logs to
// those associated with a specific commit, the commitHash value can be used. If left empty, then the search is
// explicitly restricted to logs that have no associated commit.
func (r *Repository) findEntityLogs(
	ctx context.Context, landscapeToken string, telemetryKey string, fromUnixNano uint64, toUnixNano uint64, commitHash string,
) ([]Log, error) {

	params := []any{
		clickhouse.Named("landscapeToken", landscapeToken),
		clickhouse.Named("telemetryKey", telemetryKey),
		clickhouse.Named("from", fromUnixNano),
		clickhouse.Named("to", toUnixNano),
		clickhouse.Named("commit", commitHash),
	}

	logs := []Log{}

	err := r.Conn.Select(ctx, &logs, `
		SELECT *
		FROM otel_logs
		WHERE
			ExplorvizTokenId = @landscapeToken
			AND ExplorvizVizObjectId = @telemetryKey
			AND Timestamp_ns >= @from
			AND Timestamp_ns < @to
			AND coalesce(CommitHash, '') = @commit
	`, params...)
	if err != nil {
		return []Log{}, err
	}

	return logs, nil
}
