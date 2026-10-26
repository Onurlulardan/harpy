package executor

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestGraphStageExplainsEverySiblingBeforeAnyData(t *testing.T) {
	left := graphBarrierDriver("forms", "left", "1")
	right := graphBarrierDriver("crm", "right", "1")
	barrier, err := newGraphStageBarrier(planner.HardPolicy(), []int{2})
	if err != nil {
		t.Fatal(err)
	}
	var runs atomic.Int32
	verifyExplained := func(context.Context, executionDriver) error {
		if eventIndex(left.events, "row:explain-left") < 0 ||
			eventIndex(right.events, "row:explain-right") < 0 {
			return ErrExecution
		}
		runs.Add(1)
		return nil
	}
	works := []graphStageWork{
		graphBarrierWork(0, 0, left, "left", verifyExplained),
		graphBarrierWork(0, 1, right, "right", verifyExplained),
	}
	results := submitGraphWorks(barrier, works)
	programs := &graphProgramRun{done: make(chan struct{})}
	if err := barrier.run(context.Background(), programs); err != nil {
		t.Fatal(err)
	}
	assertGraphSubmitResults(t, results, false)
	if runs.Load() != 2 {
		t.Fatalf("data runs=%d", runs.Load())
	}
}

func TestGraphStageRejectionRunsNoSiblingData(t *testing.T) {
	left := graphBarrierDriver("forms", "left", "0.6")
	right := graphBarrierDriver("crm", "right", "0.6")
	barrier, err := newGraphStageBarrier(fixtureLowPolicy(t), []int{2})
	if err != nil {
		t.Fatal(err)
	}
	var runs atomic.Int32
	run := func(context.Context, executionDriver) error {
		runs.Add(1)
		return nil
	}
	results := submitGraphWorks(barrier, []graphStageWork{
		graphBarrierWork(0, 0, left, "left", run),
		graphBarrierWork(0, 1, right, "right", run),
	})
	programs := &graphProgramRun{done: make(chan struct{})}
	if err := barrier.run(context.Background(), programs); err == nil {
		t.Fatal("cumulative sibling budget was accepted")
	}
	assertGraphSubmitResults(t, results, true)
	if runs.Load() != 0 || eventIndex(left.events, "row:explain-left") < 0 ||
		eventIndex(right.events, "row:explain-right") < 0 {
		t.Fatalf("runs=%d left=%v right=%v", runs.Load(), left.events, right.events)
	}
}

func TestGraphStageWorkRejectsEmptyAndOversizedStatementSets(t *testing.T) {
	driver := graphBarrierDriver("forms", "bounds", "1")
	work := graphStageWork{
		stage: 0, ordinal: 0, driver: driver,
		run: func(context.Context, executionDriver) error { return nil },
	}
	if work.validate() == nil {
		t.Fatal("empty unadmitted work was accepted")
	}
	work.statements = make([]statement, maximumGraphWorkStatements+1)
	for index := range work.statements {
		work.statements[index] = statement{
			sql: "data", explainSQL: "explain", class: planner.StatementClassInclude,
		}
	}
	if work.validate() == nil {
		t.Fatal("oversized work was accepted")
	}
}
