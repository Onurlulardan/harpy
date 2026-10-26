package executor

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestGraphLifecycleCleansWaitersBeforeRepanicking(t *testing.T) {
	driver := graphBarrierDriver("forms", "panic", "1")
	barrier, err := newGraphStageBarrier(planner.HardPolicy(), []int{1})
	if err != nil {
		t.Fatal(err)
	}
	programs := &graphProgramRun{done: make(chan struct{})}
	programDone := make(chan struct{})
	go func() {
		defer close(programDone)
		programs.err = barrier.submit(context.Background(), graphStageWork{
			stage: 0, ordinal: 0, driver: driver,
			statements: []statement{{
				sql: "data-panic", explainSQL: "explain-panic",
				class: planner.StatementClassInclude,
			}},
			run: func(context.Context, executionDriver) error {
				panic("driver data panic")
			},
		})
		close(programs.done)
	}()
	panicResult := make(chan any, 1)
	go func() {
		defer func() { panicResult <- recover() }()
		_, _ = runGraphLifecycle(context.Background(), barrier, programs)
	}()
	select {
	case value := <-panicResult:
		if value != "driver data panic" {
			t.Fatalf("panic=%v", value)
		}
	case <-time.After(time.Second):
		t.Fatal("lifecycle panic deadlocked")
	}
	select {
	case <-programDone:
	case <-time.After(time.Second):
		t.Fatal("waiting Program was not released")
	}
}
