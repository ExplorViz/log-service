-- Modified version of the schema from the ClickHouse exporter:
-- https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/internal/sqltemplates/logs_table.sql

CREATE TABLE IF NOT EXISTS otel_logs (
    `Timestamp` DateTime64(9) COMMENT 'Event timestamp with nanosecond precision' CODEC(Delta(8), ZSTD(1)),
    `TraceId` String COMMENT 'W3C trace identifier' CODEC(ZSTD(1)),
    `SpanId` String COMMENT 'W3C span identifier' CODEC(ZSTD(1)),
    `TraceFlags` UInt8 COMMENT 'W3C trace flags',
    `SeverityText` LowCardinality(String) COMMENT 'Log severity as text' CODEC(ZSTD(1)),
    `SeverityNumber` UInt8 COMMENT 'Log severity as number (1-24)',
    `ServiceName` LowCardinality(String) COMMENT 'Service that emitted the log' CODEC(ZSTD(1)),
    `Body` String COMMENT 'Log message body' CODEC(ZSTD(1)),
    `ResourceSchemaUrl` LowCardinality(String) COMMENT 'Schema URL for the resource' CODEC(ZSTD(1)),
    `ResourceAttributes` Map(LowCardinality(String), String) COMMENT 'Resource attributes as key-value pairs' CODEC(ZSTD(1)),
    `ScopeSchemaUrl` LowCardinality(String) COMMENT 'Schema URL for the instrumentation scope' CODEC(ZSTD(1)),
    `ScopeName` String COMMENT 'Instrumentation scope name' CODEC(ZSTD(1)),
    `ScopeVersion` LowCardinality(String) COMMENT 'Instrumentation scope version' CODEC(ZSTD(1)),
    `ScopeAttributes` Map(LowCardinality(String), String) COMMENT 'Instrumentation scope attributes' CODEC(ZSTD(1)),
    `LogAttributes` Map(LowCardinality(String), String) COMMENT 'Log record attributes' CODEC(ZSTD(1)),
    `EventName` String COMMENT 'Event name for log records representing events' CODEC(ZSTD(1)),

    -- Custom columns for ExplorViz
    `LogId` UUID DEFAULT generateUUIDv4(),

    -- Materialized columns for ExplorViz
    Timestamp_ns Int64 MATERIALIZED toUnixTimestamp64Nano(Timestamp),
    CommitHash String MATERIALIZED LogAttributes['vcs.ref.head.revision'],
    ExplorvizEntityId String MATERIALIZED LogAttributes['explorviz.entity.id'],
    ExplorvizTelemetryKey String MATERIALIZED LogAttributes['explorviz.entity.telemetrykey'],
    ExplorvizTokenId String MATERIALIZED LogAttributes['explorviz.token.id'],
    ExplorvizFuncName String MATERIALIZED LogAttributes['explorviz.code.function.name'],

    INDEX idx_trace_id TraceId TYPE text(tokenizer = 'array'),
    INDEX idx_res_attr_key mapKeys(ResourceAttributes) TYPE text(tokenizer = 'array', preprocessor = lower(mapKeys(ResourceAttributes))),
    INDEX idx_res_attr_value mapValues(ResourceAttributes) TYPE text(tokenizer = 'array', preprocessor = lower(mapValues(ResourceAttributes))),
    INDEX idx_scope_attr_key mapKeys(ScopeAttributes) TYPE text(tokenizer = 'array', preprocessor = lower(mapKeys(ScopeAttributes))),
    INDEX idx_scope_attr_value mapValues(ScopeAttributes) TYPE text(tokenizer = 'array', preprocessor = lower(mapValues(ScopeAttributes))),
    INDEX idx_log_attr_key mapKeys(LogAttributes) TYPE text(tokenizer = 'array', preprocessor = lower(mapKeys(LogAttributes))),
    INDEX idx_log_attr_value mapValues(LogAttributes) TYPE text(tokenizer = 'array', preprocessor = lower(mapValues(LogAttributes))),
    INDEX idx_lower_body Body TYPE text(tokenizer = 'splitByNonAlpha', preprocessor = lower(Body))
) ENGINE = MergeTree
PARTITION BY toDate(Timestamp)
ORDER BY (toStartOfFiveMinutes(Timestamp), ServiceName, Timestamp)
SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1