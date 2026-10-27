package deadletterdelivery

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunnerRejectsInvalidLifecycleInputs(t *testing.T) {
	settings := Settings{ReconnectMinimum: time.Second, ReconnectMaximum: 30 * time.Second}
	if _, err := New(nil, settings, openDeadLetterAdmission()); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil connector error = %v", err)
	}
	connector := newScriptedConnector(scriptedOpen{session: newBlockingSession()})
	if _, err := New(connector, settings, nil); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil admission error = %v", err)
	}
	if _, err := New(connector, Settings{}, openDeadLetterAdmission()); !errors.Is(err, ErrDependencies) {
		t.Fatalf("invalid settings error = %v", err)
	}
	runner, err := New(connector, settings, openDeadLetterAdmission())
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(nil); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil start context error = %v", err)
	}
	if err := runner.Wait(nil); !errors.Is(err, ErrDependencies) {
		t.Fatalf("nil wait context error = %v", err)
	}
	if err := runner.Force(); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerCannotStartTwiceAndForceClosesConsumer(t *testing.T) {
	session := newBlockingSession()
	connector := newScriptedConnector(scriptedOpen{session: session})
	runner, err := New(connector, Settings{
		ReconnectMinimum: time.Millisecond, ReconnectMaximum: 2 * time.Millisecond,
	}, openDeadLetterAdmission())
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-session.started:
	case <-time.After(time.Second):
		t.Fatal("consumer session did not start")
	}
	if err := runner.Start(context.Background()); !errors.Is(err, ErrAlreadyStarted) {
		t.Fatalf("second start error = %v", err)
	}
	if err := runner.Force(); err != nil {
		t.Fatal(err)
	}
	waitContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := runner.Wait(waitContext); err != nil {
		t.Fatal(err)
	}
	if session.closeCall.Load() != 1 {
		t.Fatalf("forced close count = %d", session.closeCall.Load())
	}
}

func TestRunnerReportsImpossibleSessionState(t *testing.T) {
	runner, err := New(newScriptedConnector(scriptedOpen{}), Settings{
		ReconnectMinimum: time.Millisecond, ReconnectMaximum: 2 * time.Millisecond,
	}, openDeadLetterAdmission())
	if err != nil {
		t.Fatal(err)
	}
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case failure, open := <-runner.Failures():
		if !open || !errors.Is(failure, ErrSessionUnavailable) {
			t.Fatalf("runner failure = %v, open=%t", failure, open)
		}
	case <-time.After(time.Second):
		t.Fatal("runner did not report impossible session")
	}
	if _, open := <-runner.Failures(); open {
		t.Fatal("failure channel remained open")
	}
}
