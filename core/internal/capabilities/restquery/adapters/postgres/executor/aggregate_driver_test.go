package executor

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func readyAggregateDriver(
	t *testing.T, batch consumedAggregate, rows [][]any, cost string,
) (*fakeDriver, statement) {
	t.Helper()
	candidate, err := aggregateStatement(batch.statement)
	if err != nil {
		t.Fatal(err)
	}
	driver := &fakeDriver{
		owner:     "forms",
		rows:      map[string]*fakeRows{candidate.sql: {values: rows}},
		rowValues: map[string]any{candidate.explainSQL: explainDocument(cost)},
		rowErrors: map[string]error{}, queryErrors: map[string]error{},
	}
	return driver, candidate
}

func aggregatePolicy(t *testing.T, aggregateCost float64) planner.Policy {
	t.Helper()
	limits := planner.HardLimits()
	limits.AggregateStatementCost = aggregateCost
	policy, err := planner.NewPolicy(limits)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}
