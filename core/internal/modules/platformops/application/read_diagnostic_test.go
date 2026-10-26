package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/diagnostics"
)

func TestReadDiagnosticAppliesCacheAndConsumerFailurePolicy(t *testing.T) {
	now := time.Date(2026, 7, 23, 1, 0, 0, 0, time.UTC)
	clock := &diagnosticClock{now: now}
	runner := &diagnosticRunner{fields: []diagnostics.Field{{Name: "ready", Value: "true"}}}
	reader := NewDiagnosticReader(runner, clock)
	request := api.DiagnosticRequest{ID: api.DiagnosticRecoveryReadiness,
		Consumer: api.ConsumerInteractive}
	first, err := reader.ReadDiagnostic(t.Context(), request)
	if err != nil || first.Cached || runner.calls != 1 {
		t.Fatalf("first read = %#v, calls=%d, err=%v", first, runner.calls, err)
	}
	second, err := reader.ReadDiagnostic(t.Context(), request)
	if err != nil || !second.Cached || runner.calls != 1 {
		t.Fatalf("cached read = %#v, calls=%d, err=%v", second, runner.calls, err)
	}
	runner.err = errors.New("database unavailable")
	_, err = reader.ReadDiagnostic(t.Context(), api.DiagnosticRequest{
		ID: api.DiagnosticRecoveryReadiness, Consumer: api.ConsumerBackup})
	if !errors.Is(err, diagnostics.ErrUnavailable) {
		t.Fatalf("backup consumer did not fail closed: %v", err)
	}
	clock.now = now.Add(diagnostics.CacheTTL + time.Microsecond)
	_, err = reader.ReadDiagnostic(t.Context(), request)
	if !errors.Is(err, diagnostics.ErrUnavailable) {
		t.Fatalf("expired interactive cache did not fail: %v", err)
	}
}

type diagnosticClock struct{ now time.Time }

func (clock *diagnosticClock) Now() time.Time { return clock.now }

type diagnosticRunner struct {
	fields []diagnostics.Field
	err    error
	calls  int
}

func (runner *diagnosticRunner) RunDiagnostic(
	_ context.Context,
	_ diagnostics.Definition,
) ([]diagnostics.Field, error) {
	runner.calls++
	return runner.fields, runner.err
}
