package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestPlatformOperatorMatrix(t *testing.T) {
	t.Parallel()
	exactText := mustTextType(t, api.TextCasefoldDisabled, api.TextAccentDisabled)
	foldedText := mustTextType(t, api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled)
	integerProfile, _ := api.NewIntegerProfile(-10, 10)
	integerType, _ := api.NewIntegerFieldType(integerProfile)
	decimalProfile, _ := api.NewDecimalProfile(8, 2)
	decimalType, _ := api.NewDecimalFieldType(decimalProfile)
	booleanType, _ := api.NewBooleanFieldType()
	dateType, _ := api.NewDateFieldType()
	timestampType, _ := api.NewTimestampFieldType()
	uuidType, _ := api.NewUUIDFieldType()
	relationType, _ := api.NewRelationFieldType()
	textProfile, _ := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	textListType, _ := api.NewTextListFieldType(textProfile)
	uuidListType, _ := api.NewUUIDListFieldType()
	jsonType, _ := api.NewJSONFieldType()
	tests := []struct {
		name      string
		fieldType api.FieldType
		operator  api.Operator
		allowed   bool
	}{
		{"text exact", exactText, api.OperatorBetween, true},
		{"text pattern needs profile", exactText, api.OperatorContains, false},
		{"text pattern", foldedText, api.OperatorContains, true},
		{"integer order", integerType, api.OperatorGTE, true},
		{"integer pattern", integerType, api.OperatorContains, false},
		{"decimal between", decimalType, api.OperatorBetween, true},
		{"boolean list", booleanType, api.OperatorIn, true},
		{"boolean order", booleanType, api.OperatorGT, false},
		{"date between", dateType, api.OperatorBetween, true},
		{"timestamp order", timestampType, api.OperatorLTE, true},
		{"uuid equality", uuidType, api.OperatorEq, true},
		{"uuid order", uuidType, api.OperatorLT, false},
		{"relation equality", relationType, api.OperatorEq, true},
		{"relation order", relationType, api.OperatorLT, false},
		{"unordered enum order", mustEnumType(t, false), api.OperatorLT, false},
		{"ordered enum order", mustEnumType(t, true), api.OperatorLT, true},
		{"text list collection", textListType, api.OperatorContainsAny, true},
		{"text list equality", textListType, api.OperatorEq, false},
		{"UUID list collection", uuidListType, api.OperatorContainsAll, true},
		{"UUID list IN", uuidListType, api.OperatorIn, false},
		{"JSON equality", jsonType, api.OperatorEq, true},
		{"JSON list", jsonType, api.OperatorIn, false},
		{"null universal", jsonType, api.OperatorIsNull, true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kind, err := test.fieldType.Kind()
			if err != nil {
				t.Fatal(err)
			}
			issue := validateOperator(test.fieldType, kind, test.operator)
			if (issue == "") != test.allowed {
				t.Fatalf("issue = %q, allowed = %v", issue, test.allowed)
			}
			if !test.allowed && issue != restqueryv1.IssueOperatorNotAllowed {
				t.Fatalf("issue = %q", issue)
			}
		})
	}
}
