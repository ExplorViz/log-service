package logs

type Log struct {
	// Identifier for the log. Note this is a generated ID,
	// as logs in OpenTelemetry do not have a natural ID.
	ID string `json:"id"`

	MessageBody  string `json:"messageBody"`
	TimeUnixNano int64  `json:"timeUnixNano,string"`

	TelemetryKey string `json:"telemetryKey,omitempty"`
	ServiceName  string `json:"serviceName,omitempty"`

	Severity     uint8  `json:"severity"`
	SeverityText string `json:"severityText,omitempty"`

	TraceID string `json:"traceId,omitempty"`
	SpanID  string `json:"spanId,omitempty"`

	EventName string `json:"eventName,omitempty"`

	LogAttribs      map[string]string `json:"logAttributes"`
	ResourceAttribs map[string]string `json:"resourceAttributes"`
}
