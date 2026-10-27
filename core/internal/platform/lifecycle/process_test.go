package lifecycle

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestProcessSignalCancelsStartupAndPreservesDrainEvents(t *testing.T) {
	source := make(chan os.Signal, 2)
	process := newSignalProcess(context.Background(), source, func() {})
	defer process.Stop()

	source <- os.Interrupt
	select {
	case <-process.StartupContext().Done():
	case <-time.After(time.Second):
		t.Fatal("first signal did not cancel startup")
	}
	if process.RuntimeContext().Err() != nil {
		t.Fatal("signal cancelled the runtime parent context")
	}
	if got := <-process.Interrupts(); got != os.Interrupt {
		t.Fatalf("first drain event = %v", got)
	}

	source <- os.Interrupt
	if got := <-process.Interrupts(); got != os.Interrupt {
		t.Fatalf("second drain event = %v", got)
	}
}

func TestProcessParentCancellationReachesBothContexts(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	process := newSignalProcess(parent, make(chan os.Signal), func() {})
	cancel()
	defer process.Stop()
	select {
	case <-process.StartupContext().Done():
	case <-time.After(time.Second):
		t.Fatal("parent cancellation did not reach startup")
	}
	if process.RuntimeContext().Err() == nil {
		t.Fatal("parent cancellation did not reach runtime")
	}
}
