package api

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNotificationsProjectorBoundsEveryActiveTenantWake(t *testing.T) {
	projector, realtime := projectorFixture(notificationsProjectionBatchSize, nil)
	workload, err := newNotificationsProjectorWorkload(projector, realtime)
	if err != nil {
		t.Fatal(err)
	}
	if err = workload.drainActiveTenants(t.Context()); err != nil {
		t.Fatal(err)
	}
	projector.mutex.Lock()
	defer projector.mutex.Unlock()
	want := len(realtime.tenants) * notificationsProjectionMaxBatches
	if len(projector.calls) != want {
		t.Fatalf("projection calls = %d, want %d", len(projector.calls), want)
	}
	for _, call := range projector.calls {
		if call.actor != notificationsProjectionActorID ||
			call.limit != notificationsProjectionBatchSize {
			t.Fatalf("projection authority/limit = %s/%d", call.actor, call.limit)
		}
	}
}

func TestNotificationsProjectorLifecycleDrainsRealtime(t *testing.T) {
	projector, realtime := projectorFixture(0, nil)
	waitStarted := make(chan struct{})
	wait := func(ctx context.Context, _ time.Duration) error {
		select {
		case <-waitStarted:
		default:
			close(waitStarted)
		}
		<-ctx.Done()
		return ctx.Err()
	}
	workload, err := buildNotificationsProjectorWorkload(projector, realtime, wait)
	if err != nil {
		t.Fatal(err)
	}
	if err = workload.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	<-waitStarted
	workload.BeginDrain()
	if err = workload.Wait(t.Context()); err != nil {
		t.Fatal(err)
	}
	if failure, open := <-workload.Failures(); open || failure != nil {
		t.Fatalf("failure channel = %v/%v", failure, open)
	}
	realtime.mutex.Lock()
	defer realtime.mutex.Unlock()
	if !realtime.drained || !realtime.waited {
		t.Fatalf("realtime lifecycle = drained:%v waited:%v",
			realtime.drained, realtime.waited)
	}
}

func TestNotificationsProjectorFailureIsObservable(t *testing.T) {
	want := errors.New("projection failed")
	projector, realtime := projectorFixture(0, want)
	workload, err := newNotificationsProjectorWorkload(projector, realtime)
	if err != nil {
		t.Fatal(err)
	}
	if err = workload.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	if failure := <-workload.Failures(); !errors.Is(failure, want) {
		t.Fatalf("projector failure = %v", failure)
	}
	workload.BeginDrain()
	if err = workload.Wait(t.Context()); err != nil {
		t.Fatal(err)
	}
}
