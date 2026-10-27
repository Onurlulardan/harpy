package deadletterdelivery

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

func TestRunnerClosesSessionAndWaitsWhileAdmissionPaused(t *testing.T) {
	t.Parallel()
	first := newBlockingSession()
	second := newBlockingSession()
	connector := newScriptedConnector(
		scriptedOpen{session: first}, scriptedOpen{session: second},
	)
	gate := admission.NewGate(nil)
	gate.Open()
	runner, err := New(connector, Settings{
		ReconnectMinimum: time.Millisecond,
		ReconnectMaximum: 2 * time.Millisecond,
	}, gate)
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-connector.opened:
	case <-time.After(time.Second):
		t.Fatal("dead-letter connector did not open")
	}
	waitDeadLetterSession(t, first.started)
	gate.Pause()
	select {
	case <-connector.opened:
		t.Fatal("paused admission opened another dead-letter session")
	case <-time.After(3 * admissionPollInterval):
	}
	gate.Open()
	select {
	case <-connector.opened:
	case <-time.After(time.Second):
		t.Fatal("dead-letter connector did not resume")
	}
	waitDeadLetterSession(t, second.started)
	runner.BeginDrain()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
}

func waitDeadLetterSession(t *testing.T, started <-chan struct{}) {
	t.Helper()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("dead-letter session did not start")
	}
}
