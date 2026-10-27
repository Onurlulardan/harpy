package retention

import (
	"context"
	"errors"
	"testing"
	"time"
)

type sequenceStore struct {
	results []CycleResult
	err     error
	calls   int
}

func (store *sequenceStore) RunRetentionCycle(context.Context) (CycleResult, error) {
	store.calls++
	if store.err != nil {
		return CycleResult{}, store.err
	}
	if len(store.results) == 0 {
		return NewCycleResult(false, StepCounts{})
	}
	result := store.results[0]
	store.results = store.results[1:]
	return result, nil
}

func TestRunnerWakeIsBoundedAndDetectsBacklog(t *testing.T) {
	worked, _ := NewCycleResult(true, StepCounts{JobsRedacted: 1})
	store := &sequenceStore{results: []CycleResult{worked, worked, worked, worked, worked}}
	observer := &recordingObserver{}
	runner, _ := newRunner(store, observer, openRetentionAdmission(t), waitContext, time.Now)
	backlog, err := runner.runWake(context.Background())
	if err != nil || !backlog || store.calls != maximumCyclesPerWake ||
		observer.backlogYields != 1 || len(observer.cycles) != maximumCyclesPerWake {
		t.Fatalf("bounded wake = backlog %t calls %d: %v", backlog, store.calls, err)
	}
}

func TestRunnerWakeStopsWhenWorkOrLockEnds(t *testing.T) {
	worked, _ := NewCycleResult(true, StepCounts{InboxPurged: 1})
	idle, _ := NewCycleResult(true, StepCounts{})
	for name, results := range map[string][]CycleResult{
		"idle":      {worked, idle},
		"contended": {mustCycleResult(t, false, StepCounts{})},
	} {
		t.Run(name, func(t *testing.T) {
			store := &sequenceStore{results: results}
			observer := &recordingObserver{}
			runner, _ := newRunner(
				store, observer, openRetentionAdmission(t), waitContext, time.Now,
			)
			backlog, err := runner.runWake(context.Background())
			if err != nil || backlog || store.calls != len(results) {
				t.Fatalf("wake = backlog %t calls %d: %v", backlog, store.calls, err)
			}
		})
	}
}

func TestRunnerUsesFixedBacklogAndIdleCadence(t *testing.T) {
	stop := errors.New("stop after wait")
	worked, _ := NewCycleResult(true, StepCounts{JobsPurged: 1})
	for name, fixture := range map[string]struct {
		results []CycleResult
		want    time.Duration
	}{
		"backlog": {[]CycleResult{worked, worked, worked, worked}, backlogYield},
		"idle":    {[]CycleResult{mustCycleResult(t, true, StepCounts{})}, idleInterval},
	} {
		t.Run(name, func(t *testing.T) {
			var delay time.Duration
			runner, _ := newRunner(
				&sequenceStore{results: fixture.results}, &recordingObserver{},
				openRetentionAdmission(t),
				func(_ context.Context, value time.Duration) error { delay = value; return stop },
				time.Now,
			)
			err := runner.run(context.Background())
			if !errors.Is(err, stop) || delay != fixture.want {
				t.Fatalf("cadence = %s, error %v", delay, err)
			}
		})
	}
}

func TestRunnerObservesFailedCycle(t *testing.T) {
	want := errors.New("database failed")
	observer := &recordingObserver{}
	times := []time.Time{time.Unix(10, 0), time.Unix(10, int64(25*time.Millisecond))}
	now := func() time.Time {
		value := times[0]
		times = times[1:]
		return value
	}
	runner, _ := newRunner(
		&sequenceStore{err: want}, observer, openRetentionAdmission(t), waitContext, now,
	)
	if _, err := runner.runWake(context.Background()); !errors.Is(err, want) {
		t.Fatalf("failed wake error = %v", err)
	}
	if len(observer.failures) != 1 || observer.failures[0] != 25*time.Millisecond {
		t.Fatalf("failure observations = %#v", observer.failures)
	}
}

type cancelingStore struct {
	cancel context.CancelFunc
}

func (store cancelingStore) RunRetentionCycle(context.Context) (CycleResult, error) {
	store.cancel()
	return CycleResult{}, context.Canceled
}

func TestRunnerDoesNotReportControlledCancellationAsFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	observer := &recordingObserver{}
	runner, _ := newRunner(
		cancelingStore{cancel: cancel}, observer, openRetentionAdmission(t), waitContext, time.Now,
	)
	backlog, err := runner.runWake(ctx)
	if err != nil || backlog || len(observer.failures) != 0 {
		t.Fatalf("controlled cancellation = backlog %t failures %d: %v", backlog, len(observer.failures), err)
	}
}

func mustCycleResult(t *testing.T, acquired bool, counts StepCounts) CycleResult {
	t.Helper()
	result, err := NewCycleResult(acquired, counts)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
