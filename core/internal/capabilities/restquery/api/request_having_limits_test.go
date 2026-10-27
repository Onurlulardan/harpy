package api

import "testing"

func TestRequestHavingEnforcesBooleanDepthThree(t *testing.T) {
	current := mustHavingPredicate(
		t, AggregateReferenceGroup, "status", OperatorIsNull, nil,
	)
	for depth := 1; depth <= maximumHavingDepth; depth++ {
		var err error
		current, err = NewRequestHavingNot(current)
		if err != nil {
			t.Fatalf("depth %d rejected: %v", depth, err)
		}
		if current.Depth() != depth {
			t.Fatalf("depth = %d, want %d", current.Depth(), depth)
		}
	}
	if _, err := NewRequestHavingNot(current); err == nil {
		t.Fatal("HAVING depth four accepted")
	}
}

func TestRequestHavingEnforcesTenPredicates(t *testing.T) {
	predicate := mustHavingPredicate(
		t, AggregateReferenceGroup, "status", OperatorIsNull, nil,
	)
	ten, err := NewRequestHavingAnd(repeatHaving(predicate, maximumHavingPredicates))
	if err != nil {
		t.Fatalf("ten predicates rejected: %v", err)
	}
	if ten.PredicateCount() != maximumHavingPredicates || !ten.valid() {
		t.Fatal("ten-predicate HAVING is invalid")
	}
	if _, err := NewRequestHavingAnd(repeatHaving(predicate, maximumHavingPredicates+1)); err == nil {
		t.Fatal("eleven HAVING predicates accepted")
	}
	ten.predicates--
	if ten.valid() {
		t.Fatal("tampered HAVING predicate cache validated")
	}
}

func TestRequestHavingRejectsTamperedNodeAndLiteralCaches(t *testing.T) {
	having := mustHavingWithLiteralCounts(t, []int{1, 2})
	if having.NodeCount() != 3 || !having.valid() {
		t.Fatalf("unexpected HAVING counters: %#v", having)
	}
	tamperedNodes := having
	tamperedNodes.nodes++
	if tamperedNodes.valid() {
		t.Fatal("tampered HAVING node cache validated")
	}
	tamperedLiterals := having
	tamperedLiterals.literals++
	if tamperedLiterals.valid() {
		t.Fatal("tampered HAVING literal cache validated")
	}
}

func repeatHaving(value RequestHaving, count int) []RequestHaving {
	result := make([]RequestHaving, count)
	for index := range result {
		result[index] = value
	}
	return result
}
