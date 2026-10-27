package eventdelivery

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

func TestEventRunnerClosesSessionAndWaitsWhileAdmissionIsPaused(t *testing.T) {
	policy := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	registry, _ := NewRegistry(
		fixtureCatalog(t, policy),
		[]Registration{fixtureRegistration(t, policy.Consumer, policy.ConsumerVersion)},
	)
	opens := make(chan struct{}, 3)
	connector := connectorFunction(func(context.Context, Registration) (brokerSession, error) {
		opens <- struct{}{}
		return sessionFunctions{
			run:   func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
			close: func() error { return nil },
		}, nil
	})
	gate := admission.NewGate(nil)
	if !gate.Open() {
		t.Fatal("event admission did not open")
	}
	runner, err := newRunner(registry, connector, gate, maximumJitter{}, waitContext)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitEventOpen(t, opens)
	gate.Pause()
	select {
	case <-opens:
		t.Fatal("paused admission reopened an event session")
	case <-time.After(3 * admissionPollInterval):
	}
	if !gate.Open() {
		t.Fatal("event admission did not resume")
	}
	waitEventOpen(t, opens)
	runner.BeginDrain()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
}

func waitEventOpen(t *testing.T, opens <-chan struct{}) {
	t.Helper()
	select {
	case <-opens:
	case <-time.After(time.Second):
		t.Fatal("event session did not open")
	}
}
