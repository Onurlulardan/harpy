package executor

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type executionDriver = execution.SnapshotDriver

func graphBarrierDriver(
	owner string,
	name string,
	cost string,
) *fakeDriver {
	return &fakeDriver{
		owner: owner, rows: map[string]*fakeRows{},
		rowValues: map[string]any{"explain-" + name: explainDocument(cost)},
		rowErrors: map[string]error{}, queryErrors: map[string]error{},
	}
}

func graphBarrierWork(
	stage int,
	ordinal int,
	driver execution.SnapshotDriver,
	name string,
	run func(context.Context, executionDriver) error,
) graphStageWork {
	return graphStageWork{
		stage: stage, ordinal: ordinal, driver: driver,
		statements: []statement{{
			sql: "data-" + name, explainSQL: "explain-" + name,
			arguments: []any{}, class: planner.StatementClassInclude,
		}},
		run: func(ctx context.Context, current execution.SnapshotDriver) error {
			return run(ctx, current)
		},
	}
}

func submitGraphWorks(
	barrier *graphStageBarrier,
	works []graphStageWork,
) <-chan error {
	results := make(chan error, len(works))
	for _, work := range works {
		current := work
		go func() {
			results <- barrier.submit(context.Background(), current)
		}()
	}
	return results
}

func assertGraphSubmitResults(
	t *testing.T,
	results <-chan error,
	wantError bool,
) {
	t.Helper()
	for range 2 {
		err := <-results
		if (err != nil) != wantError {
			t.Fatalf("submission error=%v wantError=%v", err, wantError)
		}
	}
}
