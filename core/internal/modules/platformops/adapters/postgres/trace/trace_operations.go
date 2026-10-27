package trace

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type Backend interface {
	TraceLookup(context.Context, string, string) ([]platformopsapi.TraceSpan, error)
	TraceListing(context.Context, string, platformopsapi.TraceFilter) (platformopsapi.TracePage, error)
	TraceRecording(context.Context, platformopsapi.TraceObservation) error
	TraceStatistics(context.Context, string) (platformopsapi.TraceStats, error)
}

type TraceOperations struct{ backend Backend }

func NewTraceOperations(backend Backend) *TraceOperations {
	return &TraceOperations{backend: backend}
}
