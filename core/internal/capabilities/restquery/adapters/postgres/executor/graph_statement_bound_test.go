package executor

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestGraphStageRejectsEighthInteractiveStatementBeforeExplain(t *testing.T) {
	barrier, err := newGraphStageBarrier(planner.HardPolicy(), []int{4})
	if err != nil {
		t.Fatal(err)
	}
	var runs atomic.Int32
	works := make([]graphStageWork, 4)
	drivers := make([]*fakeDriver, 4)
	for index := range works {
		name := string(rune('a' + index))
		drivers[index] = graphBarrierDriver("forms", name, "1")
		works[index] = graphBarrierWork(
			0, index, drivers[index], name,
			func(context.Context, executionDriver) error {
				runs.Add(1)
				return nil
			},
		)
		works[index].statements = append(
			works[index].statements,
			statement{
				sql: "data-extra-" + name, explainSQL: "explain-extra-" + name,
				class: planner.StatementClassInclude,
			},
		)
	}
	results := submitGraphWorks(barrier, works)
	programs := &graphProgramRun{done: make(chan struct{})}
	if err := barrier.run(context.Background(), programs); err == nil {
		t.Fatal("eighth interactive statement was accepted")
	}
	for range works {
		if err := <-results; err == nil {
			t.Fatal("rejected work received success")
		}
	}
	if runs.Load() != 0 {
		t.Fatalf("data runs=%d", runs.Load())
	}
	for _, driver := range drivers {
		for _, event := range driver.events {
			if len(event) >= 4 && event[:4] == "row:" {
				t.Fatalf("EXPLAIN ran before statement bound rejection: %v", driver.events)
			}
		}
	}
}
