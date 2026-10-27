package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveLiteralsRejectsInvalidTypeOperatorArityAndLiteral(t *testing.T) {
	t.Parallel()
	literal := mustTextLiteral(t, "value")
	_, err := ResolveLiterals(api.FieldType{}, api.OperatorEq, []api.Literal{literal})
	requireResolutionCode(t, err, restqueryv1.IssueTypeMismatch)

	fieldType := mustTextType(t, api.TextCasefoldDisabled, api.TextAccentDisabled)
	_, err = ResolveLiterals(fieldType, api.Operator(255), []api.Literal{literal})
	requireResolutionCode(t, err, restqueryv1.IssueOperatorNotAllowed)
	_, err = ResolveLiterals(fieldType, api.OperatorEq, nil)
	requireResolutionCode(t, err, restqueryv1.IssueArityMismatch)
	_, err = ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{{}})
	requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
}

func TestResolveNullOperatorsReturnCanonicalEmptyList(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewJSONFieldType()
	values, err := ResolveLiterals(fieldType, api.OperatorIsNull, nil)
	if err != nil {
		t.Fatal(err)
	}
	if values == nil || len(values) != 0 {
		t.Fatalf("null operator values = %#v", values)
	}
	_, err = ResolveLiterals(
		fieldType, api.OperatorIsNull, []api.Literal{mustTextLiteral(t, "value")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueArityMismatch)
}

func TestResolveDisallowedOperatorPrecedesLiteralInspection(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewBooleanFieldType()
	_, err := ResolveLiterals(fieldType, api.OperatorContains, []api.Literal{{}})
	requireResolutionCode(t, err, restqueryv1.IssueOperatorNotAllowed)
	_, err = ResolveLiterals(fieldType, api.OperatorContains, nil)
	requireResolutionCode(t, err, restqueryv1.IssueOperatorNotAllowed)
}
