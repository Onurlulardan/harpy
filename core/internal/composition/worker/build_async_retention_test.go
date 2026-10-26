package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type retentionBuildObserver struct{}

func (retentionBuildObserver) ObserveCycle(context.Context, retention.CycleObservation) {}
func (retentionBuildObserver) ObserveFailure(context.Context, time.Duration)            {}
func (retentionBuildObserver) ObserveBacklogYield(context.Context)                      {}

func TestBuildAsyncRetentionRequiresWorkerPoolAndObserver(t *testing.T) {
	gate := admission.NewGate(nil)
	if runner, err := buildAsyncRetention(nil, retentionBuildObserver{}, gate); runner != nil ||
		!errors.Is(err, platformpostgres.ErrPoolClosed) {
		t.Fatalf("nil retention pool = %#v, %v", runner, err)
	}
	if runner, err := buildAsyncRetention(nil, nil, gate); runner != nil ||
		!errors.Is(err, platformpostgres.ErrPoolClosed) {
		t.Fatalf("nil retention dependencies = %#v, %v", runner, err)
	}
}
