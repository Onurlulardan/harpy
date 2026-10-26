package api

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	usagemodule "github.com/Onurlulardan/harpy/core/internal/modules/usage"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestattribution"
)

type usageRequestRecorderFixture struct {
	request usagemodule.RecordRequest
	calls   int
	err     error
}

func (fixture *usageRequestRecorderFixture) Record(
	_ context.Context, request usagemodule.RecordRequest,
) (usagemodule.RecordResult, error) {
	fixture.calls++
	fixture.request = request
	return usagemodule.RecordResult{}, fixture.err
}

func TestUsageRequestObserverRecordsVerifiedCanonicalTraffic(t *testing.T) {
	recorder := &usageRequestRecorderFixture{}
	environment := [16]byte{1}
	observer, err := newUsageRequestObserver(environment, recorder)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Date(2026, 8, 1, 10, 11, 12, 123456000, time.UTC)
	observation := httpedge.RequestObservation{
		SourceIdentity: [32]byte{9},
		Attribution: requestattribution.Identity{
			ActorID: [16]byte{2}, TenantID: [16]byte{3},
		},
		AttributionState: requestattribution.StatusVerified,
		Method:           "GET", RoutePattern: "/api/analytics/dashboards/{id}",
		StatusCode: 200, StartedAt: started, EndedAt: started.Add(1500 * time.Microsecond),
	}
	if err := observer(context.Background(), observation); err != nil {
		t.Fatal(err)
	}
	request := recorder.request
	if recorder.calls != 1 || request.EnvironmentID != environment ||
		request.ActorID != observation.Attribution.ActorID ||
		request.TenantID != observation.Attribution.TenantID ||
		request.SourceIdentity != observation.SourceIdentity ||
		request.RoutePattern != observation.RoutePattern || request.DurationMicros != 1500 ||
		!request.OccurredAt.Equal(started) {
		t.Fatalf("Usage request mapping drifted: %#v", request)
	}
}

func TestUsageRequestObserverExcludesOnlyUsageSurface(t *testing.T) {
	recorder := &usageRequestRecorderFixture{}
	observer, err := newUsageRequestObserver([16]byte{1}, recorder)
	if err != nil {
		t.Fatal(err)
	}
	base := httpedge.RequestObservation{SourceIdentity: [32]byte{1},
		Attribution:      requestattribution.Identity{ActorID: [16]byte{2}, TenantID: [16]byte{3}},
		AttributionState: requestattribution.StatusVerified, Method: "GET", StatusCode: 200,
		StartedAt: time.Now().UTC().Truncate(time.Microsecond)}
	base.EndedAt = base.StartedAt
	for _, route := range []string{"/api/usage-analytics", "/api/usage-analytics/overview"} {
		base.RoutePattern = route
		if err := observer(context.Background(), base); err != nil {
			t.Fatal(err)
		}
	}
	base.RoutePattern = "/api/analytics/overview"
	if err := observer(context.Background(), base); err != nil {
		t.Fatal(err)
	}
	if recorder.calls != 1 {
		t.Fatalf("record calls = %d", recorder.calls)
	}
}

func TestUsageRequestObserverSkipsUnattributedAndRejectsConflict(t *testing.T) {
	recorder := &usageRequestRecorderFixture{}
	observer, err := newUsageRequestObserver([16]byte{1}, recorder)
	if err != nil {
		t.Fatal(err)
	}
	observation := httpedge.RequestObservation{RoutePattern: "/api/forms"}
	observation.AttributionState = requestattribution.StatusUnattributed
	if err := observer(context.Background(), observation); err != nil {
		t.Fatal(err)
	}
	observation.AttributionState = requestattribution.StatusConflict
	if err := observer(context.Background(), observation); !errors.Is(err, errUsageObservationConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	if recorder.calls != 0 {
		t.Fatalf("untrusted attribution recorded %d facts", recorder.calls)
	}
}

func TestParallelRequestObserverDoesNotStarveSibling(t *testing.T) {
	failure := errors.New("first observer failed")
	var first, second atomic.Bool
	observer, err := newParallelRequestObserver(
		func(context.Context, httpedge.RequestObservation) error {
			first.Store(true)
			return failure
		},
		func(context.Context, httpedge.RequestObservation) error {
			second.Store(true)
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := observer(context.Background(), httpedge.RequestObservation{}); !errors.Is(err, failure) || !first.Load() || !second.Load() {
		t.Fatalf("parallel observation = %v, first=%v second=%v", err, first.Load(), second.Load())
	}
}
