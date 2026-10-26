package platformops

import (
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type spanTreeView struct {
	SpanID        string         `json:"span_id"`
	ServiceName   string         `json:"service_name"`
	Operation     string         `json:"operation"`
	SpanKind      string         `json:"span_kind"`
	DurationMS    float64        `json:"duration_ms"`
	StatusCode    int            `json:"status_code,omitempty"`
	Level         string         `json:"level"`
	HasError      bool           `json:"has_error"`
	StartOffsetMS float64        `json:"start_offset_ms,omitempty"`
	Children      []spanTreeView `json:"children"`
}

type traceDetailView struct {
	CorrelationID   string          `json:"correlation_id"`
	Spans           []traceSpanView `json:"spans"`
	Tree            spanTreeView    `json:"tree"`
	TotalSpans      int             `json:"total_spans"`
	TotalDurationMS float64         `json:"total_duration_ms"`
}

type timelineView struct {
	CorrelationID string             `json:"correlation_id"`
	BaseTime      time.Time          `json:"base_time"`
	Timeline      []timelineSpanWire `json:"timeline"`
}

type timelineSpanWire struct {
	SpanID        string  `json:"span_id"`
	ParentSpanID  string  `json:"parent_span_id,omitempty"`
	ServiceName   string  `json:"service_name"`
	Operation     string  `json:"operation"`
	SpanKind      string  `json:"span_kind"`
	StartOffsetMS float64 `json:"start_offset_ms"`
	DurationMS    float64 `json:"duration_ms"`
	Level         string  `json:"level"`
	StatusCode    int     `json:"status_code,omitempty"`
	HasError      bool    `json:"has_error"`
}

func traceSpanAPI(value platformopsapi.TraceSpan) traceSpanView {
	return traceSpanView{ID: value.SpanID, CorrelationID: value.CorrelationID,
		SpanID: value.SpanID, ParentSpanID: value.ParentSpanID,
		ServiceName: value.ServiceName, Operation: value.Operation,
		SpanKind: value.SpanKind, Method: value.Method, Path: value.Path,
		StatusCode: value.StatusCode, StartTime: value.StartedAt, EndTime: value.EndedAt,
		DurationMS: value.DurationMillis, Level: value.Level, Message: value.Message,
		CreatedAt: value.CreatedAt}
}
