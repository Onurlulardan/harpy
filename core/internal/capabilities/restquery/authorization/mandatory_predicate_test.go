package authorization

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestMandatoryPredicateIsTypedImmutableAndRedacted(t *testing.T) {
	predicate := mandatoryTextPredicate(t)
	fieldID, fieldErr := predicate.FieldID()
	kind, kindErr := predicate.LogicalType()
	operator, operatorErr := predicate.Operator()
	values, valuesErr := predicate.Values()
	if fieldErr != nil || fieldID.String() != testFieldID ||
		kindErr != nil || kind != api.LogicalTypeText ||
		operatorErr != nil || operator != api.OperatorEq ||
		valuesErr != nil || len(values) != 1 {
		t.Fatal("mandatory predicate accessors are invalid")
	}
	values[0] = resolved.Value{}
	if predicate.Validate() != nil {
		t.Fatal("returned values mutated mandatory predicate")
	}
	if fmt.Sprintf("%v", predicate) != authorizationRedaction ||
		predicate.LogValue().String() != authorizationRedaction {
		t.Fatal("mandatory predicate was not redacted")
	}
	if _, err := json.Marshal(predicate); err == nil {
		t.Fatal("mandatory predicate allowed direct JSON marshal")
	}
}

func TestMandatoryPredicateRejectsTypeOperatorAndIdentifierMismatch(t *testing.T) {
	boolean, err := resolved.NewBooleanValue(true)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		fieldID string
		kind    api.LogicalType
		op      api.Operator
		values  []resolved.Value
	}{
		{strings.ToUpper("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), api.LogicalTypeText, api.OperatorEq,
			[]resolved.Value{mandatoryScalar(t, api.LogicalTypeText, "x")}},
		{testFieldID, api.LogicalTypeBoolean, api.OperatorGT, []resolved.Value{boolean}},
		{testFieldID, api.LogicalTypeText, api.OperatorEq, []resolved.Value{boolean}},
	}
	for index, test := range cases {
		if _, err := NewMandatoryPredicate(
			test.fieldID, test.kind, test.op, test.values,
		); err == nil {
			t.Fatalf("invalid mandatory predicate %d was accepted", index)
		}
	}
}
