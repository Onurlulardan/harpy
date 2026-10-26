package api

import (
	"encoding/json"
	"testing"
)

func TestOperatorCanonicalSetAndArity(t *testing.T) {
	tests := []struct {
		name  string
		arity PredicateArity
	}{
		{"eq", PredicateArityOne}, {"ne", PredicateArityOne},
		{"gt", PredicateArityOne}, {"gte", PredicateArityOne},
		{"lt", PredicateArityOne}, {"lte", PredicateArityOne},
		{"contains", PredicateArityOne}, {"starts_with", PredicateArityOne},
		{"ends_with", PredicateArityOne}, {"in", PredicateArityList},
		{"not_in", PredicateArityList}, {"between", PredicateArityPair},
		{"is_null", PredicateArityZero}, {"is_not_null", PredicateArityZero},
		{"contains_any", PredicateArityList}, {"contains_all", PredicateArityList},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			operator, found := ParseOperator(test.name)
			if !found || operator.String() != test.name {
				t.Fatalf("operator did not round trip: %q", test.name)
			}
			arity, found := operator.Arity()
			if !found || arity != test.arity {
				t.Fatalf("unexpected arity for %s: %v", test.name, arity)
			}
		})
	}
	if len(tests) != len(canonicalOperators()) {
		t.Fatalf("canonical operator count differs: %d", len(canonicalOperators()))
	}
}

func TestOperatorRejectsUnknownValue(t *testing.T) {
	if operator, found := ParseOperator("equals"); found || operator != 0 {
		t.Fatalf("unknown operator accepted: %v", operator)
	}
	if Operator(255).String() != "" {
		t.Fatal("unknown operator has a wire name")
	}
	if _, found := Operator(255).Arity(); found {
		t.Fatal("unknown operator has an arity")
	}
}

func TestFilterEnumsCannotMarshalDirectly(t *testing.T) {
	for _, value := range []any{OperatorEq, PredicateArityOne, RequestFilterPredicate} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}
