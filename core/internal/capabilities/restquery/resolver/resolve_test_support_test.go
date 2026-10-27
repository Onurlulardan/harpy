package resolver

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

const resolverEnumVersion = "22222222-2222-4222-8222-222222222222"

func mustJSONLiteral(t *testing.T, document string) api.Literal {
	t.Helper()
	literal, err := api.NewJSONLiteral([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func mustTextLiteral(t *testing.T, text string) api.Literal {
	t.Helper()
	literal, err := api.NewTextLiteral(text)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func requireResolutionCode(t *testing.T, err error, expected restqueryv1.IssueCode) {
	t.Helper()
	if err == nil {
		t.Fatal("resolution error = nil")
	}
	var resolutionError LiteralResolutionError
	if !errors.As(err, &resolutionError) {
		t.Fatalf("resolution error type = %T", err)
	}
	code, codeErr := resolutionError.Code()
	if codeErr != nil || code != expected {
		t.Fatalf("resolution code = %q, err = %v, want %q", code, codeErr, expected)
	}
}

func requireScalars(t *testing.T, values []resolved.Value, expected ...string) {
	t.Helper()
	if len(values) != len(expected) {
		t.Fatalf("value count = %d, want %d", len(values), len(expected))
	}
	for index, expectedValue := range expected {
		actual, found, err := values[index].Scalar()
		if err != nil || !found || actual != expectedValue {
			t.Fatalf("value %d scalar = %q, found = %v, err = %v", index, actual, found, err)
		}
	}
}

func mustTextType(
	t *testing.T, casefold api.TextCasefoldProfile, accent api.TextAccentProfile,
) api.FieldType {
	t.Helper()
	profile, err := api.NewTextProfile(casefold, accent)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewTextFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func mustEnumType(t *testing.T, ordered bool) api.FieldType {
	t.Helper()
	var first, second *int
	if ordered {
		zero, one := 0, 1
		first, second = &zero, &one
	}
	open, _ := api.NewEnumValue("OPEN", first)
	closed, _ := api.NewEnumValue("CLOSED", second)
	profile, err := api.NewEnumProfile(resolverEnumVersion, []api.EnumValue{open, closed})
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewEnumFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}
