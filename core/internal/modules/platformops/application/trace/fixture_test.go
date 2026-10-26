package trace

import (
	"context"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

type backendFixture struct{ operation string }

func (backend *backendFixture) TraceLookup(
	context.Context, string, string,
) ([]platformopsapi.TraceSpan, error) {
	backend.operation = "get"
	return []platformopsapi.TraceSpan{{}}, nil
}

func (backend *backendFixture) TraceListing(
	context.Context, string, platformopsapi.TraceFilter,
) (platformopsapi.TracePage, error) {
	backend.operation = "list"
	return platformopsapi.TracePage{Total: 1}, nil
}

func (backend *backendFixture) TraceRecording(
	context.Context, platformopsapi.TraceObservation,
) error {
	backend.operation = "record"
	return nil
}

func (backend *backendFixture) TraceStatistics(
	context.Context, string,
) (platformopsapi.TraceStats, error) {
	backend.operation = "read"
	return platformopsapi.TraceStats{TotalTraces: 1}, nil
}
