package planner

import (
	"errors"
	"math"
	"testing"
	"time"
)

func TestHardPolicyStatementContracts(t *testing.T) {
	t.Parallel()
	policy := HardPolicy()
	tests := []struct {
		class   StatementClass
		name    string
		cost    float64
		timeout time.Duration
		shared  bool
	}{
		{StatementClassRecord, "record", 50_000, 5 * time.Second, true},
		{StatementClassInclude, "include", 50_000, 5 * time.Second, true},
		{StatementClassV1PredecessorSeek, "v1_predecessor_seek", 50_000, 5 * time.Second, true},
		{StatementClassExactCount, "exact_count", 50_000, 2 * time.Second, false},
		{StatementClassAggregate, "aggregate", 100_000, 5 * time.Second, false},
	}
	for _, test := range tests {
		statement, err := policy.Statement(test.class)
		if err != nil {
			t.Fatalf("statement %s: %v", test.name, err)
		}
		if test.class.String() != test.name || statement.Class() != test.class ||
			statement.MaximumCost() != test.cost || statement.Timeout() != test.timeout ||
			statement.SharesRequestBudget() != test.shared {
			t.Fatalf("statement %s contract = %#v", test.name, statement)
		}
	}
}

func TestPolicyMayOnlyLowerHardLimits(t *testing.T) {
	t.Parallel()
	lowered := Limits{
		RecordStatementCost: 25_000, RecordRequestCost: 75_000,
		ExactCountStatementCost: 20_000, AggregateStatementCost: 80_000,
		InteractiveTimeout: 3 * time.Second, ExactCountTimeout: time.Second,
	}
	policy, err := NewPolicy(lowered)
	if err != nil || policy.Limits() != lowered {
		t.Fatalf("lowered policy = %#v, %v", policy, err)
	}
	invalid := []Limits{}
	for _, mutate := range []func(*Limits){
		func(value *Limits) { value.RecordStatementCost = 50_001 },
		func(value *Limits) { value.RecordRequestCost = 100_001 },
		func(value *Limits) { value.ExactCountStatementCost = 50_001 },
		func(value *Limits) { value.AggregateStatementCost = 100_001 },
		func(value *Limits) { value.InteractiveTimeout = 5*time.Second + 1 },
		func(value *Limits) { value.ExactCountTimeout = 2*time.Second + 1 },
		func(value *Limits) { value.RecordStatementCost = 0 },
		func(value *Limits) { value.RecordStatementCost = math.NaN() },
	} {
		candidate := HardLimits()
		mutate(&candidate)
		invalid = append(invalid, candidate)
	}
	for _, candidate := range invalid {
		if _, err := NewPolicy(candidate); !errors.Is(err, ErrPolicyInvalid) {
			t.Fatalf("invalid policy error = %v", err)
		}
	}
}

func TestPolicyCopiesLimits(t *testing.T) {
	t.Parallel()
	input := HardLimits()
	policy, _ := NewPolicy(input)
	input.RecordRequestCost = 1
	output := policy.Limits()
	output.RecordRequestCost = 2
	if policy.Limits().RecordRequestCost != HardRecordRequestCost {
		t.Fatal("policy limits were mutated through a copy")
	}
}
