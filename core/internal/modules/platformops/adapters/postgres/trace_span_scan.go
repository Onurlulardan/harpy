package postgres

import platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"

type traceSpanRow interface{ Scan(...any) error }

func scanTraceSpan(row traceSpanRow) (platformopsapi.TraceSpan, error) {
	var item platformopsapi.TraceSpan
	var id string
	var parent, message *string
	if err := row.Scan(&id, &item.EnvironmentID, &item.CorrelationID,
		&item.SpanID, &parent, &item.ServiceName, &item.Operation,
		&item.SpanKind, &item.Method, &item.Path, &item.StatusCode,
		&item.StartedAt, &item.EndedAt, &item.DurationMillis,
		&item.Level, &message, &item.CreatedAt); err != nil {
		return platformopsapi.TraceSpan{}, err
	}
	if parent != nil {
		item.ParentSpanID = *parent
	}
	if message != nil {
		item.Message = *message
	}
	return item, nil
}
