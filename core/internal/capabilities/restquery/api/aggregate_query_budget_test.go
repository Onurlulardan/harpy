package api

import "testing"

func TestAggregateQueryCombinesRootAndHavingLiteralBudget(t *testing.T) {
	root := mustPredicate(t)
	exactHaving := mustHavingWithLiteralCounts(t,
		[]int{100, 100, 100, 100, 100, 100, 100, 100, 100, 99})
	exactAggregate := mustAggregateWithHaving(t, exactHaving)
	if _, err := NewAggregateQueryRequest(AggregateQueryRequestInput{
		Where: &root, Aggregate: exactAggregate,
	}); err != nil {
		t.Fatalf("1000 combined literals rejected: %v", err)
	}
	overHaving := mustHavingWithLiteralCounts(t,
		[]int{100, 100, 100, 100, 100, 100, 100, 100, 100, 100})
	overAggregate := mustAggregateWithHaving(t, overHaving)
	if _, err := NewAggregateQueryRequest(AggregateQueryRequestInput{
		Where: &root, Aggregate: overAggregate,
	}); err == nil {
		t.Fatal("1001 combined literals accepted")
	}
}

func TestAggregateQueryCombinesRootAndHavingNodeBudget(t *testing.T) {
	having := mustTwentyNodeHaving(t)
	aggregate := mustAggregateWithHaving(t, having)
	exactRoot := mustRootFilterNodes(t, 39, true)
	if _, err := NewAggregateQueryRequest(AggregateQueryRequestInput{
		Where: &exactRoot, Aggregate: aggregate,
	}); err != nil {
		t.Fatalf("100 combined nodes rejected: %v", err)
	}
	overRoot := mustRootFilterNodes(t, 40, false)
	if _, err := NewAggregateQueryRequest(AggregateQueryRequestInput{
		Where: &overRoot, Aggregate: aggregate,
	}); err == nil {
		t.Fatal("101 combined nodes accepted")
	}
}

func mustHavingWithLiteralCounts(t *testing.T, counts []int) RequestHaving {
	t.Helper()
	literal := mustLiteral(t, `"value"`)
	leaves := make([]RequestHaving, len(counts))
	for index, count := range counts {
		values := make([]Literal, count)
		for valueIndex := range values {
			values[valueIndex] = literal
		}
		leaves[index] = mustHavingPredicate(
			t, AggregateReferenceGroup, "status", OperatorIn, values,
		)
	}
	having, err := NewRequestHavingAnd(leaves)
	if err != nil {
		t.Fatal(err)
	}
	return having
}

func mustTwentyNodeHaving(t *testing.T) RequestHaving {
	t.Helper()
	leaf := mustHavingPredicate(t, AggregateReferenceGroup, "status", OperatorIsNull, nil)
	children := make([]RequestHaving, 9)
	for index := range children {
		children[index], _ = NewRequestHavingNot(leaf)
	}
	children[0], _ = NewRequestHavingNot(children[0])
	having, err := NewRequestHavingAnd(children)
	if err != nil {
		t.Fatal(err)
	}
	return having
}

func mustRootFilterNodes(t *testing.T, predicates int, doubleFirst bool) RequestFilter {
	t.Helper()
	children := wrappedPredicates(t, mustPredicate(t), predicates)
	if doubleFirst {
		children[0], _ = NewRequestNot(children[0])
	}
	filter, err := NewRequestAnd(children)
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

func mustAggregateWithHaving(t *testing.T, having RequestHaving) RequestAggregate {
	t.Helper()
	aggregate, err := NewRequestAggregate(RequestAggregateInput{
		GroupBy: []string{"status"}, Having: &having,
	})
	if err != nil {
		t.Fatal(err)
	}
	return aggregate
}
