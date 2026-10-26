package published

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type expectedPublishedIssue struct {
	path string
	code restqueryv1.IssueCode
}

func requirePublishedError(
	t *testing.T, err error, expected []expectedPublishedIssue,
) api.QueryError {
	t.Helper()
	if err == nil {
		t.Fatal("error = nil")
	}
	var queryError api.QueryError
	if !errors.As(err, &queryError) {
		t.Fatalf("error type = %T, want api.QueryError", err)
	}
	if queryError.Code() != restqueryv1.ErrorQueryParameterInvalid {
		t.Fatalf("code = %q", queryError.Code())
	}
	issues := queryError.Issues()
	if len(issues) != len(expected) {
		t.Fatalf("issues = %#v, want %d", issues, len(expected))
	}
	for index, want := range expected {
		if issues[index].Path != want.path || issues[index].Code != string(want.code) {
			t.Fatalf("issue %d = %#v, want %#v", index, issues[index], want)
		}
	}
	return queryError
}

func testIntegerFieldType(t *testing.T, minimum, maximum int64) api.FieldType {
	t.Helper()
	profile, err := api.NewIntegerProfile(minimum, maximum)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewIntegerFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func testTextFieldType(t *testing.T) api.FieldType {
	t.Helper()
	profile, err := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewTextFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func testDecimalFieldType(t *testing.T, precision, scale int) api.FieldType {
	t.Helper()
	profile, err := api.NewDecimalProfile(precision, scale)
	if err != nil {
		t.Fatal(err)
	}
	fieldType, err := api.NewDecimalFieldType(profile)
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func testBooleanFieldType(t *testing.T) api.FieldType {
	t.Helper()
	fieldType, err := api.NewBooleanFieldType()
	if err != nil {
		t.Fatal(err)
	}
	return fieldType
}

func mustScalarValue(
	t *testing.T, kind api.LogicalType, value string,
) resolved.Value {
	t.Helper()
	result, err := resolved.NewScalarValue(kind, value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func mustTextLiteral(t *testing.T, value string) api.Literal {
	t.Helper()
	literal, err := api.NewTextLiteral(value)
	if err != nil {
		t.Fatal(err)
	}
	return literal
}

func mustBooleanValue(t *testing.T, value bool) resolved.Value {
	t.Helper()
	result, err := resolved.NewBooleanValue(value)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
