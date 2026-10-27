package eventdelivery

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRunnerReconnectsWithBoundedJitterAndResetsAfterOpen(t *testing.T) {
	policy := fixturePolicy(t, "workflow.projector", 1, "forms.entry.created")
	registry, _ := NewRegistry(
		fixtureCatalog(t, policy),
		[]Registration{fixtureRegistration(t, policy.Consumer, policy.ConsumerVersion)},
	)
	mutex := &sync.Mutex{}
	opens, closed := 0, 0
	blocking := make(chan string, 1)
	connector := connectorFunction(func(context.Context, Registration) (brokerSession, error) {
		mutex.Lock()
		opens++
		attempt := opens
		mutex.Unlock()
		if attempt <= 2 {
			return nil, errors.New("broker unavailable")
		}
		if attempt == 3 {
			return sessionFunctions{
				run: func(context.Context) error { return errors.New("stream closed") },
				close: func() error {
					mutex.Lock()
					closed++
					mutex.Unlock()
					return nil
				},
			}, nil
		}
		return sessionFunctions{
			run: func(ctx context.Context) error {
				blocking <- "connected"
				<-ctx.Done()
				return ctx.Err()
			},
			close: func() error { return nil },
		}, nil
	})
	waits := &recordedWait{}
	runner, _ := newRunner(
		registry, connector, openEventAdmission(t), maximumJitter{}, waits.call,
	)
	_ = runner.Start(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := waitSignal(ctx, blocking); err != nil {
		t.Fatal(err)
	}
	runner.BeginDrain()
	if err := runner.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	delays := waits.values()
	want := []time.Duration{time.Second, 2 * time.Second, time.Second}
	if len(delays) != len(want) {
		t.Fatalf("unexpected reconnect waits: %v", delays)
	}
	for index := range want {
		if delays[index] != want[index] {
			t.Fatalf("reconnect wait %d: got %s want %s", index, delays[index], want[index])
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	if opens != 4 || closed != 1 {
		t.Fatalf("unexpected session lifecycle: opens=%d closed=%d", opens, closed)
	}
}
