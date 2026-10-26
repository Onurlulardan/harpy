package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveEnumUsesExactCatalogLookup(t *testing.T) {
	t.Parallel()
	fieldType := mustEnumType(t, false)
	values, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `"OPEN"`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "OPEN")
	_, err = ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, "open")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueValueNotAllowed)
	_, err = ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `true`)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueTypeMismatch)
}

func TestResolveEnumOrderRequiresOrderedProfile(t *testing.T) {
	t.Parallel()
	_, err := ResolveLiterals(
		mustEnumType(t, false), api.OperatorGT,
		[]api.Literal{mustTextLiteral(t, "OPEN")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueOperatorNotAllowed)
	values, err := ResolveLiterals(
		mustEnumType(t, true), api.OperatorGT,
		[]api.Literal{mustTextLiteral(t, "OPEN")},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "OPEN")
}

func TestResolveJSONCanonicalizesPOSTAndGET(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewJSONFieldType()
	for _, literal := range []api.Literal{
		mustJSONLiteral(t, `{"b":2,"a":1}`),
		mustTextLiteral(t, `{"b":2,"a":1}`),
	} {
		values, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{literal})
		if err != nil {
			t.Fatal(err)
		}
		document, found, accessorErr := values[0].JSON()
		if accessorErr != nil || !found || string(document) != `{"a":1,"b":2}` {
			t.Fatalf("canonical JSON = %q, found = %v, err = %v", document, found, accessorErr)
		}
	}
}

func TestResolveJSONRejectsNULAtEveryDepth(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewJSONFieldType()
	for _, literal := range []api.Literal{
		mustJSONLiteral(t, `{"nested":{"value":"\u0000"}}`),
		mustTextLiteral(t, `{"nested":{"bad\u0000key":true}}`),
	} {
		_, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{literal})
		requireResolutionCode(t, err, restqueryv1.IssueNULForbidden)
	}
}

func TestResolveJSONRejectsNonCanonicalizableDocuments(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewJSONFieldType()
	for _, literal := range []api.Literal{
		mustTextLiteral(t, `null`),
		mustTextLiteral(t, `{"a":1,"a":2}`),
		mustJSONLiteral(t, `1e1000`),
		mustTextLiteral(t, `not-json`),
	} {
		_, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{literal})
		requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
	}
}
