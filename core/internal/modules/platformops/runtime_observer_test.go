package platformops

import (
	"bytes"
	"context"
	"testing"
	"time"

	platformopsapi "github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
)

type runtimeObserverEnvironment struct{}

func (runtimeObserverEnvironment) Current(context.Context) (HTTPEnvironment, error) {
	return HTTPEnvironment{
		EnvironmentID: "018f47a2-7b7c-7c9d-8a2d-3c4b5e6f7081",
		TenantID:      "018f47a2-7b7c-7c9d-8a2d-3c4b5e6f7082", DeploymentMode: "onpremise",
	}, nil
}

type runtimeObserverRecorder struct {
	observation platformopsapi.TraceObservation
	calls       int
}

func (recorder *runtimeObserverRecorder) RecordTrace(
	_ context.Context,
	observation platformopsapi.TraceObservation,
) error {
	recorder.observation = observation
	recorder.calls++
	return nil
}

func TestRuntimeObserverRecordsBoundedRequestTrace(t *testing.T) {
	recorder := &runtimeObserverRecorder{}
	observer, err := newRuntimeTraceObserver(runtimeObserverEnvironment{}, recorder)
	if err != nil {
		t.Fatal(err)
	}
	observer.entropy = bytes.NewReader([]byte("12345678"))
	started := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	err = observer.Observe(t.Context(), httpedge.RequestObservation{
		CorrelationID: requestmeta.CorrelationID("phase14-observer"), Method: "GET",
		RoutePattern: "/api/forms", StatusCode: 503,
		StartedAt: started, EndedAt: started.Add(5 * time.Millisecond),
	})
	if err != nil || recorder.calls != 1 || recorder.observation.Level != "error" ||
		recorder.observation.SpanID != "3132333435363738" {
		t.Fatalf("observation = %#v, calls=%d, err=%v", recorder.observation, recorder.calls, err)
	}
}

func TestRuntimeObserverExcludesOperationsSurface(t *testing.T) {
	recorder := &runtimeObserverRecorder{}
	observer, err := newRuntimeTraceObserver(runtimeObserverEnvironment{}, recorder)
	if err != nil {
		t.Fatal(err)
	}
	err = observer.Observe(t.Context(), httpedge.RequestObservation{
		CorrelationID: requestmeta.CorrelationID("phase14-self"), Method: "GET",
		RoutePattern: "/api/platform/operations/health", StatusCode: 200,
		StartedAt: time.Now().UTC(), EndedAt: time.Now().UTC(),
	})
	if err != nil || recorder.calls != 0 {
		t.Fatalf("self observation calls=%d, err=%v", recorder.calls, err)
	}
}
