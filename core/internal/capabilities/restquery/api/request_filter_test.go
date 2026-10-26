package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestRequestPredicateEnforcesOperatorArity(t *testing.T) {
	value := mustLiteral(t, `"value"`)
	tests := []struct {
		operator Operator
		count    int
		valid    bool
	}{
		{OperatorIsNull, 0, true}, {OperatorIsNull, 1, false},
		{OperatorEq, 1, true}, {OperatorEq, 0, false}, {OperatorEq, 2, false},
		{OperatorIn, 1, true}, {OperatorIn, 100, true},
		{OperatorIn, 0, false}, {OperatorIn, 101, false},
		{OperatorBetween, 2, true}, {OperatorBetween, 1, false},
	}
	for _, test := range tests {
		values := make([]Literal, test.count)
		for index := range values {
			values[index] = value
		}
		_, err := NewRequestPredicate("field", test.operator, values)
		if (err == nil) != test.valid {
			t.Fatalf("operator=%s count=%d valid=%v err=%v",
				test.operator, test.count, test.valid, err)
		}
	}
	if _, err := NewRequestPredicate("field", Operator(255), nil); err == nil {
		t.Fatal("unknown operator accepted")
	}
}

func TestRequestFilterIsImmutableRecursiveUnion(t *testing.T) {
	literal := mustLiteral(t, `"secret"`)
	predicate, err := NewRequestPredicate("field", OperatorEq, []Literal{literal})
	if err != nil {
		t.Fatal(err)
	}
	children := []RequestFilter{predicate}
	conjunction, err := NewRequestAnd(children)
	if err != nil {
		t.Fatal(err)
	}
	children[0] = RequestFilter{}
	returned := conjunction.Children()
	returned[0] = RequestFilter{}
	stored := conjunction.Children()[0]
	value, found := stored.Predicate()
	if !found || value.Field() != "field" || value.Operator() != OperatorEq {
		t.Fatalf("predicate changed: %#v", value)
	}
	values := value.Values()
	values[0] = Literal{}
	if !storedValuesRemainValid(stored) {
		t.Fatal("predicate exposed mutable literal storage")
	}
}

func TestRequestFilterKindsAndFormatting(t *testing.T) {
	predicate := mustPredicate(t)
	conjunction, _ := NewRequestAnd([]RequestFilter{predicate})
	disjunction, _ := NewRequestOr([]RequestFilter{predicate})
	negation, _ := NewRequestNot(predicate)
	if predicate.Kind() != RequestFilterPredicate || conjunction.Kind() != RequestFilterAnd ||
		disjunction.Kind() != RequestFilterOr || negation.Kind() != RequestFilterNot {
		t.Fatal("filter kind differs")
	}
	value, _ := predicate.Predicate()
	for _, formatted := range []string{
		fmt.Sprint(predicate), fmt.Sprintf("%#v", predicate), fmt.Sprintf("%q", predicate),
		fmt.Sprint(value), fmt.Sprintf("%#v", value),
	} {
		if formatted != queryValueRedaction {
			t.Fatalf("filter formatting leaked value: %q", formatted)
		}
	}
}

func TestRequestASTCannotBeDirectlyJSONMarshaled(t *testing.T) {
	secret := mustLiteral(t, `"do-not-leak"`)
	filter, err := NewRequestPredicate("field", OperatorEq, []Literal{secret})
	if err != nil {
		t.Fatal(err)
	}
	predicate, _ := filter.Predicate()
	for name, value := range map[string]any{
		"filter": filter, "predicate": predicate,
	} {
		_, err := json.Marshal(value)
		if err == nil {
			t.Fatalf("%s was silently marshaled", name)
		}
		if strings.Contains(err.Error(), "do-not-leak") {
			t.Fatalf("%s marshal error leaked value: %v", name, err)
		}
	}
}

func TestRequestPredicateRejectsInvalidUTF8AndLiteral(t *testing.T) {
	for _, field := range []string{"", "field\x00name", "e\u0301", string([]byte{0xff})} {
		if _, err := NewRequestPredicate(field, OperatorIsNull, nil); err == nil {
			t.Fatalf("invalid field %q accepted", field)
		}
	}
	if _, err := NewRequestPredicate("", OperatorEq, []Literal{{}}); err == nil {
		t.Fatal("invalid literal accepted")
	}
}

func mustLiteral(t *testing.T, value string) Literal {
	t.Helper()
	literal, err := NewJSONLiteral([]byte(value))
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func mustPredicate(t *testing.T) RequestFilter {
	t.Helper()
	filter, err := NewRequestPredicate("field", OperatorEq, []Literal{mustLiteral(t, `"value"`)})
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

func storedValuesRemainValid(filter RequestFilter) bool {
	predicate, found := filter.Predicate()
	return found && len(predicate.Values()) == 1 && predicate.Values()[0].valid()
}
