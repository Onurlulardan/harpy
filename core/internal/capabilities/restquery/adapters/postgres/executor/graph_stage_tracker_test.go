package executor

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestGraphStagesShareOneImmutableRequestTracker(t *testing.T) {
	driver := graphBarrierDriver("forms", "shared", "0.6")
	driver.rowValues["explain-next"] = explainDocument("0.6")
	barrier, err := newGraphStageBarrier(fixtureLowPolicy(t), []int{1, 1})
	if err != nil {
		t.Fatal(err)
	}
	var first, second atomic.Int32
	results := submitGraphWorks(barrier, []graphStageWork{
		graphBarrierWork(0, 0, driver, "shared", func(
			context.Context, executionDriver,
		) error {
			first.Add(1)
			return nil
		}),
		graphBarrierWork(1, 0, driver, "next", func(
			context.Context, executionDriver,
		) error {
			second.Add(1)
			return nil
		}),
	})
	programs := &graphProgramRun{done: make(chan struct{})}
	if err := barrier.run(context.Background(), programs); err == nil {
		t.Fatal("second stage reset the request tracker")
	}
	nilCount, errorCount := 0, 0
	for range 2 {
		if err := <-results; err == nil {
			nilCount++
		} else {
			errorCount++
		}
	}
	if nilCount != 1 || errorCount != 1 {
		t.Fatalf("submission results nil=%d error=%d", nilCount, errorCount)
	}
	if first.Load() != 1 || second.Load() != 0 {
		t.Fatalf("first=%d second=%d", first.Load(), second.Load())
	}
}

func TestGraphStageWaitOnlyParticipantCannotRunData(t *testing.T) {
	barrier, err := newGraphStageBarrier(fixtureLowPolicy(t), []int{1})
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- barrier.wait(context.Background(), 0, 0) }()
	programs := &graphProgramRun{done: make(chan struct{})}
	if err := barrier.run(context.Background(), programs); err != nil {
		t.Fatal(err)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
}
