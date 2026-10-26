package platformops

import platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"

type traceStatsView struct {
	TotalTraces       int64                          `json:"total_traces"`
	TotalSpans        int64                          `json:"total_spans"`
	ErrorRate         float64                        `json:"error_rate"`
	AverageDurationMS float64                        `json:"avg_duration_ms"`
	SpansByService    map[string]int64               `json:"spans_by_service"`
	SpansByLevel      map[string]int64               `json:"spans_by_level"`
	TracesLast24H     int64                          `json:"traces_last_24h"`
	TracesLastHour    int64                          `json:"traces_last_hour"`
	SlowestEndpoints  []platformopsapi.EndpointStats `json:"slowest_endpoints"`
}

func traceStatsAPI(value platformopsapi.TraceStats) traceStatsView {
	return traceStatsView{TotalTraces: value.TotalTraces, TotalSpans: value.TotalSpans,
		ErrorRate: value.ErrorRate, AverageDurationMS: value.AverageDurationMillis,
		SpansByService: value.SpansByService, SpansByLevel: value.SpansByLevel,
		TracesLast24H: value.TracesLast24Hours, TracesLastHour: value.TracesLastHour,
		SlowestEndpoints: value.SlowestEndpoints}
}
