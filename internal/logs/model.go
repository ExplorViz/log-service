package logs

type Log struct {
	MessageBody  string `json:"messageBody"`
	TimeUnixNano uint64 `json:"timeUnixNano,string"`

	TraceId   string `json:"traceId,omitempty"`
	SpanId    string `json:"spanId,omitempty"`
	Severity  string `json:"severity,omitempty"`
	EventName string `json:"eventName,omitempty"`

	LogAttribs      map[string]string `json:"logAttributes"`
	ResourceAttribs map[string]string `json:"resourceAttributes"`
}
