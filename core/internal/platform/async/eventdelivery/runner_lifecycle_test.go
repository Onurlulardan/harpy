package eventdelivery

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRunnerStartsOneSessionPerIdentityAndDrains(t *testing.T) {
	first := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	second := fixturePolicy(t, "analytics.projector", 1, "forms.entry.created")
	registry, _ := NewRegistry(fixtureCatalog(t, first, second), []Registration{
		fixtureRegistration(t, first.Consumer, first.ConsumerVersion),
		fixtureRegistration(t, second.Consumer, second.ConsumerVersion),
	})
	started := make(chan string, 2)
	connector := connectorFunction(func(
		_ context.Context, registration Registration,
	) (brokerSession, error) {
		return sessionFunctions{
			run: func(ctx context.Context) error {
				started <- registration.Consumer()
				<-ctx.Done()
				return ctx.Err()
			},
			close: func() error { return nil },
		}, nil
	})
	runner, _ := newRunner(
		registry, connector, openEventAdmission(t), maximumJitter{}, waitContext,
	)
	if err := runner.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	seen := map[string]bool{}
	for range 2 {
		identity, err := waitSignal(ctx, started)
		if err != nil {
			t.Fatal(err)
		}
		seen[identity] = true
	}
	if !seen[first.Consumer] || !seen[second.Consumer] {
		t.Fatalf("identity sessions were not concurrent: %v", seen)
	}
	runner.BeginDrain()
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if _, open := <-runner.Failures(); open {
		t.Fatal("clean drain left failure stream open")
	}
}

func TestRunnerForceClosesSessionThatIgnoresCancellation(t *testing.T) {
	policy := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	registry, _ := NewRegistry(
		fixtureCatalog(t, policy),
		[]Registration{fixtureRegistration(t, policy.Consumer, policy.ConsumerVersion)},
	)
	started, closed := make(chan string, 1), make(chan struct{})
	closeOnce := &sync.Once{}
	connector := connectorFunction(func(context.Context, Registration) (brokerSession, error) {
		return sessionFunctions{
			run: func(context.Context) error {
				started <- "started"
				<-closed
				return errors.New("closed")
			},
			close: func() error { closeOnce.Do(func() { close(closed) }); return nil },
		}, nil
	})
	runner, _ := newRunner(
		registry, connector, openEventAdmission(t), maximumJitter{}, waitContext,
	)
	_ = runner.Start(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := waitSignal(ctx, started); err != nil {
		t.Fatal(err)
	}
	runner.BeginDrain()
	short, stop := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer stop()
	if !errors.Is(runner.Wait(short), context.DeadlineExceeded) {
		t.Fatal("drain unexpectedly completed while session ignored cancellation")
	}
	if err := runner.Force(); err != nil {
		t.Fatal(err)
	}
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerReportsPanickedSession(t *testing.T) {
	policy := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	registry, _ := NewRegistry(
		fixtureCatalog(t, policy),
		[]Registration{fixtureRegistration(t, policy.Consumer, policy.ConsumerVersion)},
	)
	connector := connectorFunction(func(context.Context, Registration) (brokerSession, error) {
		return sessionFunctions{
			run:   func(context.Context) error { panic("provider detail") },
			close: func() error { return nil },
		}, nil
	})
	runner, _ := newRunner(
		registry, connector, openEventAdmission(t), maximumJitter{}, waitContext,
	)
	_ = runner.Start(context.Background())
	failure := <-runner.Failures()
	if !errors.Is(failure, ErrSessionPanicked) {
		t.Fatalf("panic was not converted to managed failure: %v", failure)
	}
	if err := runner.Wait(context.Background()); err != nil {
		t.Fatal(err)
	}
}
