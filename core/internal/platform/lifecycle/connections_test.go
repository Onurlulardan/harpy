package lifecycle

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestConnectionsCancelCloseAndWaitForRelease(t *testing.T) {
	connections := NewConnections()
	var closes atomic.Int32
	tracked, release, err := connections.Track(context.Background(), func() error {
		closes.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("track: %v", err)
	}
	if err := connections.BeginDrain(); err != nil {
		t.Fatalf("begin drain: %v", err)
	}
	select {
	case <-tracked.Done():
	default:
		t.Fatal("tracked context was not cancelled")
	}
	if closes.Load() != 1 {
		t.Fatalf("close count = %d", closes.Load())
	}
	short, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := connections.Wait(short); err == nil {
		t.Fatal("expected wait to block until release")
	}
	release()
	if err := connections.Wait(context.Background()); err != nil {
		t.Fatalf("wait after release: %v", err)
	}
	if err := connections.Force(); err != nil {
		t.Fatalf("force: %v", err)
	}
	if closes.Load() != 1 {
		t.Fatalf("close must be idempotent, count = %d", closes.Load())
	}
}

func TestConnectionsRejectTrackingAfterDrain(t *testing.T) {
	connections := NewConnections()
	if err := connections.BeginDrain(); err != nil {
		t.Fatalf("begin drain: %v", err)
	}
	if _, _, err := connections.Track(context.Background(), nil); err == nil {
		t.Fatal("expected closed-admission error")
	}
}
