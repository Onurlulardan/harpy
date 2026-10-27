package api

import "testing"

func TestRequestFilterEnforcesChildAndDepthLimits(t *testing.T) {
	predicate := mustPredicate(t)
	if _, err := NewRequestAnd(nil); err == nil {
		t.Fatal("empty boolean accepted")
	}
	if _, err := NewRequestAnd(repeatFilters(predicate, 51)); err == nil {
		t.Fatal("boolean with 51 children accepted")
	}
	current := predicate
	if current.Depth() != 0 {
		t.Fatalf("predicate depth = %d, want 0", current.Depth())
	}
	for depth := 1; depth <= 5; depth++ {
		var err error
		current, err = NewRequestNot(current)
		if err != nil {
			t.Fatalf("depth %d rejected: %v", depth, err)
		}
		if current.Depth() != depth {
			t.Fatalf("unexpected depth: %d", current.Depth())
		}
	}
	if _, err := NewRequestNot(current); err == nil {
		t.Fatal("depth 6 accepted")
	}
}

func TestRequestFilterRejectsMalformedRecursiveShape(t *testing.T) {
	predicate := mustPredicate(t)
	badArity := &RequestPredicate{field: "field", operator: OperatorEq}
	malformed := []RequestFilter{
		{kind: RequestFilterAnd, depth: 1, nodes: 1},
		{kind: RequestFilterNot, children: repeatFilters(predicate, 2), depth: 1, nodes: 3},
		{kind: RequestFilterPredicate, predicate: badArity, nodes: 1},
		{kind: RequestFilterAnd, children: []RequestFilter{predicate}, depth: 1, nodes: 99},
		{
			kind: RequestFilterAnd,
			children: []RequestFilter{{
				kind: RequestFilterAnd, children: nil, depth: 1, nodes: 1,
			}},
			depth: 2, nodes: 2,
		},
	}
	for index, child := range malformed {
		if _, err := NewRequestAnd([]RequestFilter{child}); err == nil {
			t.Fatalf("malformed child %d accepted", index)
		}
	}
}

func TestRequestFilterEnforcesNodeLimit(t *testing.T) {
	predicate := mustPredicate(t)
	exactChildren := wrappedPredicates(t, predicate, 49)
	exactChildren[0], _ = NewRequestNot(exactChildren[0])
	exact, err := NewRequestAnd(exactChildren)
	if err != nil {
		t.Fatalf("100 nodes rejected: %v", err)
	}
	if exact.NodeCount() != 100 {
		t.Fatalf("unexpected node count: %d", exact.NodeCount())
	}
	if _, err := NewRequestAnd(wrappedPredicates(t, predicate, 50)); err == nil {
		t.Fatal("101 nodes accepted")
	}
}

func TestRequestFilterEnforcesPredicateLimit(t *testing.T) {
	predicate := mustPredicate(t)
	inner, err := NewRequestAnd(repeatFilters(predicate, maximumFilterPredicates))
	if err != nil {
		t.Fatalf("50 predicates rejected: %v", err)
	}
	if err := inner.Validate(); err != nil {
		t.Fatalf("constructed filter failed validation: %v", err)
	}
	if _, err := NewRequestAnd([]RequestFilter{inner, predicate}); err == nil {
		t.Fatal("51 predicates accepted")
	}
	inner.predicates--
	if err := inner.Validate(); err == nil {
		t.Fatal("tampered predicate cache validated")
	}
}

func repeatFilters(filter RequestFilter, count int) []RequestFilter {
	filters := make([]RequestFilter, count)
	for index := range filters {
		filters[index] = filter
	}
	return filters
}

func wrappedPredicates(t *testing.T, predicate RequestFilter, count int) []RequestFilter {
	t.Helper()
	filters := make([]RequestFilter, count)
	for index := range filters {
		wrapped, err := NewRequestNot(predicate)
		if err != nil {
			t.Fatal(err)
		}
		filters[index] = wrapped
	}
	return filters
}
