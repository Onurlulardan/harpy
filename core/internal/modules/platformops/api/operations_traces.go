package api

import (
	"context"
	"time"
)

type TraceObservation struct {
	EnvironmentID, CorrelationID, SpanID, ParentSpanID string
	ServiceName, Operation, SpanKind, Method, Path     string
	StatusCode                                         int
	StartedAt, EndedAt                                 time.Time
	DurationMillis                                     float64
	Level, Message                                     string
}

type TraceFilter struct {
	Page, Limit                           int
	ServiceName, Operation, Level, Search string
	StatusCode, MinStatus, MaxStatus      int
	MinimumDuration, MaximumDuration      float64
	StartedAfter, EndedBefore             *time.Time
	SortBy, SortDirection                 string
}

type TraceSummary struct {
	CorrelationID, RootOperation string
	FirstSpan, LastSpan          time.Time
	TotalDurationMillis          float64
	SpanCount, RootStatusCode    int
	Services                     []string
	HasErrors                    bool
}

type TraceSpan struct {
	TraceObservation
	CreatedAt time.Time
}

type TracePage struct {
	Traces                  []TraceSummary
	Total                   int64
	Page, Limit, TotalPages int
}

type EndpointStats struct {
	Path                  string  `json:"path"`
	Method                string  `json:"method"`
	AverageDurationMillis float64 `json:"avg_duration_ms"`
	CallCount             int64   `json:"call_count"`
	ErrorCount            int64   `json:"error_count"`
}

type TraceStats struct {
	TotalTraces, TotalSpans, TracesLast24Hours, TracesLastHour int64
	ErrorRate, AverageDurationMillis                           float64
	SpansByService, SpansByLevel                               map[string]int64
	SlowestEndpoints                                           []EndpointStats
}

type TraceOperations interface {
	RecordTrace(context.Context, TraceObservation) error
	ListTraces(context.Context, string, TraceFilter) (TracePage, error)
	GetTrace(context.Context, string, string) ([]TraceSpan, error)
	GetTraceSpan(context.Context, string, string) (TraceSpan, error)
	ReadTraceStats(context.Context, string) (TraceStats, error)
}
