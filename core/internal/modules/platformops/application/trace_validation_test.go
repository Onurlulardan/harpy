package application

import (
	"math"
	"testing"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
)

func TestTraceObservationAcceptsOnlyBoundedRedactedShape(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	valid := platformopsapi.TraceObservation{
		EnvironmentID: "018f0000-0000-7000-8000-000000000001",
		CorrelationID: "client_request-1", SpanID: "0123456789abcdef",
		ServiceName: "core-api", Operation: "GET /api/users/{user_id}",
		SpanKind: "server", Method: "GET", Path: "/api/users/{user_id}",
		StatusCode: 200, StartedAt: now, EndedAt: now.Add(time.Millisecond),
		DurationMillis: 1, Level: "info",
	}
	if !validTraceObservation(valid) {
		t.Fatal("valid trace was rejected")
	}
	tests := []func(*platformopsapi.TraceObservation){
		func(value *platformopsapi.TraceObservation) { value.CorrelationID = "contains secret?" },
		func(value *platformopsapi.TraceObservation) { value.SpanID = "not-hex" },
		func(value *platformopsapi.TraceObservation) { value.Path = "/raw?token=secret" },
		func(value *platformopsapi.TraceObservation) { value.StatusCode = 999 },
		func(value *platformopsapi.TraceObservation) { value.DurationMillis = math.NaN() },
		func(value *platformopsapi.TraceObservation) { value.Level = "verbose" },
	}
	for index, mutate := range tests {
		candidate := valid
		mutate(&candidate)
		if validTraceObservation(candidate) {
			t.Fatalf("invalid trace %d was accepted", index)
		}
	}
}

func TestTraceFilterRejectsInvalidNumericAndStatusRanges(t *testing.T) {
	valid := platformopsapi.TraceFilter{Page: 1, Limit: 50,
		SortBy: "started_at", SortDirection: "desc"}
	if !validTraceFilter(valid) {
		t.Fatal("valid filter was rejected")
	}
	for _, mutate := range []func(*platformopsapi.TraceFilter){
		func(value *platformopsapi.TraceFilter) { value.Level = "verbose" },
		func(value *platformopsapi.TraceFilter) { value.StatusCode = 99 },
		func(value *platformopsapi.TraceFilter) { value.MinStatus, value.MaxStatus = 500, 400 },
		func(value *platformopsapi.TraceFilter) { value.MinimumDuration = math.Inf(1) },
	} {
		candidate := valid
		mutate(&candidate)
		if validTraceFilter(candidate) {
			t.Fatalf("invalid trace filter was accepted: %#v", candidate)
		}
	}
}
