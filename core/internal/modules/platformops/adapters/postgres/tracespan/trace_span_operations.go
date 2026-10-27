package tracespan

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	TraceSpanLookup(context.Context, string, string) (platformopsapi.TraceSpan, error)
}

type TraceSpanOperations struct{ backend Backend }

func NewTraceSpanOperations(backend Backend) *TraceSpanOperations {
	return &TraceSpanOperations{backend: backend}
}
