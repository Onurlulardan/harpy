package retention

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

func TestRetentionWakeDoesNotTouchStoreWhileAdmissionIsClosed(t *testing.T) {
	store := &sequenceStore{}
	gate := admission.NewGate(nil)
	runner, err := newRunner(
		store, &recordingObserver{}, gate, waitContext, time.Now,
	)
	if err != nil {
		t.Fatal(err)
	}
	backlog, err := runner.runWake(context.Background())
	if err != nil || backlog || store.calls != 0 {
		t.Fatalf("closed admission backlog=%t calls=%d err=%v", backlog, store.calls, err)
	}
	if !gate.Open() {
		t.Fatal("retention admission did not open")
	}
	if _, err := runner.runWake(context.Background()); err != nil || store.calls != 1 {
		t.Fatalf("opened admission calls=%d err=%v", store.calls, err)
	}
}
