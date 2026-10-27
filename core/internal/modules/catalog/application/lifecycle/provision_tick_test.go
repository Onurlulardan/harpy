package lifecycle

import (
	"testing"
	"time"
)

func TestProvisionTickContinuesPastPendingOwner(t *testing.T) {
	started := time.Date(2026, 8, 19, 14, 0, 0, 0, time.UTC)
	retryAt := started.Add(time.Second)
	wait, stop := provisionTickDecision(false, retryAt, started, started, true)
	if stop || !wait.Equal(retryAt) {
		t.Fatalf("pending owner blocked later owners wait=%v stop=%v", wait, stop)
	}
}

func TestProvisionTickYieldsAfterBudgetWhenOwnersRemain(t *testing.T) {
	started := time.Date(2026, 8, 19, 14, 0, 0, 0, time.UTC)
	now := started.Add(applyTickBudget)
	wait, stop := provisionTickDecision(true, time.Time{}, started, now, true)
	if !stop || !wait.IsZero() {
		t.Fatalf("slow ready owner kept the lease wait=%v stop=%v", wait, stop)
	}
}

func TestProvisionTickKeepsFastReadyOwnersOnTheSameJob(t *testing.T) {
	started := time.Date(2026, 8, 19, 14, 0, 0, 0, time.UTC)
	now := started.Add(applyTickBudget - time.Second)
	wait, stop := provisionTickDecision(true, time.Time{}, started, now, true)
	if stop || !wait.IsZero() {
		t.Fatalf("fast ready owner yielded wait=%v stop=%v", wait, stop)
	}
	wait, stop = provisionTickDecision(true, time.Time{}, started, now, false)
	if stop || !wait.IsZero() {
		t.Fatalf("final ready owner stopped wait=%v stop=%v", wait, stop)
	}
}

func TestFloorProvisionRetryRejectsStalePastRetry(t *testing.T) {
	now := time.Date(2026, 8, 19, 14, 30, 0, 0, time.UTC)
	stale := now.Add(-time.Hour)
	got := floorProvisionRetry(stale, now)
	want := now.Add(time.Second)
	if !got.Equal(want) {
		t.Fatalf("stale retry at=%v want=%v", got, want)
	}
	if !floorProvisionRetry(time.Time{}, now).IsZero() {
		t.Fatal("ready tick was delayed")
	}
}

func TestEarliestProvisionWaitKeepsSoonestPendingOwner(t *testing.T) {
	first := time.Date(2026, 8, 19, 14, 0, 2, 0, time.UTC)
	second := time.Date(2026, 8, 19, 14, 0, 1, 0, time.UTC)
	if got := earliestProvisionWait(first, second); !got.Equal(second) {
		t.Fatalf("earliest wait=%v", got)
	}
}
