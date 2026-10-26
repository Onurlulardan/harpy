package planner

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestTrackerStatementLimitsAndErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		class StatementClass
		limit float64
		code  restqueryv1.ErrorCode
	}{
		{StatementClassRecord, 50_000, restqueryv1.ErrorQueryBudgetExceeded},
		{StatementClassInclude, 50_000, restqueryv1.ErrorQueryBudgetExceeded},
		{StatementClassV1PredecessorSeek, 50_000, restqueryv1.ErrorQueryBudgetExceeded},
		{StatementClassExactCount, 50_000, restqueryv1.ErrorQueryCountBudgetExceeded},
		{StatementClassAggregate, 100_000, restqueryv1.ErrorQueryBudgetExceeded},
	}
	for _, test := range tests {
		tracker, _ := NewTracker(HardPolicy())
		if _, err := tracker.Admit(test.class, mustCost(t, test.limit)); err != nil {
			t.Fatalf("class %s boundary: %v", test.class, err)
		}
		unchanged, err := tracker.Admit(test.class, mustCost(t, test.limit+0.01))
		requireQueryError(t, err, test.code, true)
		if unchanged.RecordRequestCost().Value() != 0 {
			t.Fatalf("failed admission changed tracker: %s", unchanged.RecordRequestCost())
		}
	}
}

func TestTrackerSharesRecordRequestBudget(t *testing.T) {
	t.Parallel()
	original, _ := NewTracker(HardPolicy())
	record, err := original.Admit(StatementClassRecord, mustCost(t, 40_000))
	if err != nil {
		t.Fatal(err)
	}
	include, err := record.Admit(StatementClassInclude, mustCost(t, 30_000))
	if err != nil {
		t.Fatal(err)
	}
	seek, err := include.Admit(StatementClassV1PredecessorSeek, mustCost(t, 30_000))
	if err != nil || seek.RecordRequestCost().Value() != 100_000 {
		t.Fatalf("boundary tracker = %s, %v", seek.RecordRequestCost(), err)
	}
	if original.RecordRequestCost().Value() != 0 ||
		record.RecordRequestCost().Value() != 40_000 ||
		include.RecordRequestCost().Value() != 70_000 {
		t.Fatal("an earlier immutable tracker value changed")
	}
	unchanged, err := seek.Admit(StatementClassInclude, mustCost(t, 0.01))
	requireQueryError(t, err, restqueryv1.ErrorQueryBudgetExceeded, true)
	if unchanged.RecordRequestCost().Value() != 100_000 {
		t.Fatal("cumulative rejection changed tracker")
	}
}

func TestSeparateStatementsDoNotConsumeRecordRequestBudget(t *testing.T) {
	t.Parallel()
	tracker, _ := NewTracker(HardPolicy())
	for _, class := range []StatementClass{
		StatementClassExactCount, StatementClassAggregate,
	} {
		statement, _ := HardPolicy().Statement(class)
		var err error
		tracker, err = tracker.Admit(class, mustCost(t, statement.MaximumCost()))
		if err != nil {
			t.Fatal(err)
		}
	}
	if tracker.RecordRequestCost().Value() != 0 {
		t.Fatalf("record request cost = %s", tracker.RecordRequestCost())
	}
}

func mustCost(t testing.TB, value float64) Cost {
	t.Helper()
	cost, err := NewCost(value)
	if err != nil {
		t.Fatal(err)
	}
	return cost
}
