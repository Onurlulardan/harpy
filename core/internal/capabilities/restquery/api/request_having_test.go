package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRequestHavingPredicateValidatesReferenceAndArity(t *testing.T) {
	literal := mustLiteral(t, `"value"`)
	group := mustHavingPredicate(t, AggregateReferenceGroup, "status", OperatorEq, []Literal{literal})
	measure := mustHavingPredicate(t, AggregateReferenceMeasure, "row_count", OperatorBetween,
		[]Literal{literal, literal})
	for _, value := range []RequestHaving{group, measure} {
		predicate, found := value.Predicate()
		if !found || predicate.Reference() == 0 || predicate.Name() == "" {
			t.Fatalf("unexpected HAVING predicate: %#v", predicate)
		}
	}
	for _, test := range []struct {
		reference AggregateReference
		name      string
		operator  Operator
		values    []Literal
	}{
		{AggregateReferenceGroup, "", OperatorIsNull, nil},
		{AggregateReferenceGroup, "e\u0301", OperatorIsNull, nil},
		{AggregateReferenceMeasure, "Upper", OperatorIsNull, nil},
		{AggregateReferenceMeasure, "valid", OperatorBetween, []Literal{literal}},
		{AggregateReference(99), "valid", OperatorIsNull, nil},
	} {
		if _, err := NewRequestHavingPredicate(
			test.reference, test.name, test.operator, test.values,
		); err == nil {
			t.Fatalf("invalid HAVING predicate accepted: %#v", test)
		}
	}
}

func TestRequestHavingIsImmutableRedactedAndCannotMarshal(t *testing.T) {
	literal := mustLiteral(t, `"secret"`)
	values := []Literal{literal}
	predicate := mustHavingPredicate(
		t, AggregateReferenceGroup, "secret_group", OperatorEq, values,
	)
	children := []RequestHaving{predicate}
	having, err := NewRequestHavingAnd(children)
	if err != nil {
		t.Fatal(err)
	}
	values[0] = Literal{}
	children[0] = RequestHaving{}
	returned := having.Children()
	returned[0] = RequestHaving{}
	stored, found := having.Children()[0].Predicate()
	if !found || len(stored.Values()) != 1 || !stored.Values()[0].valid() {
		t.Fatal("HAVING exposed mutable storage")
	}
	for _, value := range []string{
		fmt.Sprint(having), fmt.Sprintf("%#v", having), fmt.Sprint(stored),
	} {
		if value != queryValueRedaction {
			t.Fatalf("HAVING formatting leaked: %q", value)
		}
	}
	for _, value := range []any{having, stored, having.Kind()} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}

func mustHavingPredicate(
	t *testing.T, reference AggregateReference, name string,
	operator Operator, values []Literal,
) RequestHaving {
	t.Helper()
	having, err := NewRequestHavingPredicate(reference, name, operator, values)
	if err != nil {
		t.Fatal(err)
	}
	return having
}
