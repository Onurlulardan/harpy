package platformops

import platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"

func traceTimelineAPI(spans []platformopsapi.TraceSpan) timelineView {
	base := spans[0].StartedAt
	for _, span := range spans[1:] {
		if span.StartedAt.Before(base) {
			base = span.StartedAt
		}
	}
	result := timelineView{CorrelationID: spans[0].CorrelationID,
		BaseTime: base, Timeline: make([]timelineSpanWire, len(spans))}
	for index, span := range spans {
		result.Timeline[index] = timelineSpanWire{SpanID: span.SpanID,
			ParentSpanID: span.ParentSpanID, ServiceName: span.ServiceName,
			Operation: span.Operation, SpanKind: span.SpanKind,
			StartOffsetMS: float64(span.StartedAt.Sub(base).Microseconds()) / 1000,
			DurationMS:    span.DurationMillis, Level: span.Level,
			StatusCode: span.StatusCode,
			HasError:   span.StatusCode >= 500 || span.Level == "error" || span.Level == "fatal"}
	}
	return result
}
