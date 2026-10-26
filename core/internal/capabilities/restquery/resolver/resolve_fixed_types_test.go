package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveBooleanSupportsPOSTJSONAndGETText(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewBooleanFieldType()
	for _, literal := range []api.Literal{
		mustJSONLiteral(t, `true`), mustTextLiteral(t, "true"),
	} {
		values, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{literal})
		if err != nil {
			t.Fatal(err)
		}
		actual, found, accessorErr := values[0].Boolean()
		if accessorErr != nil || !found || !actual {
			t.Fatalf("boolean = %v, found = %v, err = %v", actual, found, accessorErr)
		}
	}
	_, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `"true"`)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueTypeMismatch)
	_, err = ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, "TRUE")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
}

func TestResolveTemporalTypesRequireExactWireForms(t *testing.T) {
	t.Parallel()
	dateType, _ := api.NewDateFieldType()
	values, err := ResolveLiterals(
		dateType, api.OperatorEq, []api.Literal{mustTextLiteral(t, "2026-07-15")},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "2026-07-15")
	_, err = ResolveLiterals(
		dateType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `20260715`)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueTypeMismatch)
	_, err = ResolveLiterals(
		dateType, api.OperatorEq, []api.Literal{mustTextLiteral(t, "2026-02-30")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)

	timestampType, _ := api.NewTimestampFieldType()
	canonical := "2026-07-15T10:11:12.123456Z"
	values, err = ResolveLiterals(
		timestampType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `"`+canonical+`"`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, canonical)
	_, err = ResolveLiterals(
		timestampType, api.OperatorEq, []api.Literal{mustTextLiteral(t, "2026-07-15T10:11:12Z")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
}

func TestResolveUUIDRequiresLowercaseNonzeroRFC4122(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewUUIDFieldType()
	canonical := "550e8400-e29b-41d4-a716-446655440000"
	values, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `"`+canonical+`"`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, canonical)
	for _, literal := range []string{
		"550E8400-E29B-41D4-A716-446655440000",
		"00000000-0000-0000-0000-000000000000",
		"550e8400-e29b-41d4-7716-446655440000",
	} {
		_, err = ResolveLiterals(
			fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, literal)},
		)
		requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
	}
}

func TestResolveRelationAndUUIDListKeepElementKinds(t *testing.T) {
	t.Parallel()
	canonical := "550e8400-e29b-41d4-a716-446655440000"
	relationType, _ := api.NewRelationFieldType()
	values, err := ResolveLiterals(
		relationType, api.OperatorEq, []api.Literal{mustTextLiteral(t, canonical)},
	)
	if err != nil {
		t.Fatal(err)
	}
	kind, _ := values[0].Kind()
	if kind != api.LogicalTypeRelation {
		t.Fatalf("relation kind = %v", kind)
	}
	uuidListType, _ := api.NewUUIDListFieldType()
	values, err = ResolveLiterals(
		uuidListType, api.OperatorContainsAny, []api.Literal{mustTextLiteral(t, canonical)},
	)
	if err != nil {
		t.Fatal(err)
	}
	kind, _ = values[0].Kind()
	if kind != api.LogicalTypeUUID {
		t.Fatalf("UUID list element kind = %v", kind)
	}
}
