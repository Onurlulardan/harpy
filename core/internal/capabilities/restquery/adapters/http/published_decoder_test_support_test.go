package httpquery

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type expectedPublishedHTTPIssue struct {
	path string
	code restqueryv1.IssueCode
}

func requirePublishedHTTPError(
	t *testing.T, err error, code restqueryv1.ErrorCode,
	expected []expectedPublishedHTTPIssue,
) api.QueryError {
	t.Helper()
	if err == nil {
		t.Fatal("error = nil")
	}
	var queryError api.QueryError
	if !errors.As(err, &queryError) {
		t.Fatalf("error type = %T, want api.QueryError", err)
	}
	if queryError.Code() != code {
		t.Fatalf("code = %q, want %q", queryError.Code(), code)
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

func publishedTestManifest(t *testing.T) published.ParameterManifest {
	t.Helper()
	integerProfile, err := api.NewIntegerProfile(0, 10)
	if err != nil {
		t.Fatal(err)
	}
	integerType, err := api.NewIntegerFieldType(integerProfile)
	if err != nil {
		t.Fatal(err)
	}
	textProfile, err := api.NewTextProfile(
		api.TextCasefoldDisabled, api.TextAccentDisabled,
	)
	if err != nil {
		t.Fatal(err)
	}
	textType, err := api.NewTextFieldType(textProfile)
	if err != nil {
		t.Fatal(err)
	}
	booleanType, err := api.NewBooleanFieldType()
	if err != nil {
		t.Fatal(err)
	}
	jsonType, err := api.NewJSONFieldType()
	if err != nil {
		t.Fatal(err)
	}
	fallback, err := resolved.NewScalarValue(api.LogicalTypeText, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := published.NewParameterManifest([]published.ParameterInput{
		{Name: "count", Type: integerType,
			Arity: published.ParameterAritySingle, Required: true},
		{Name: "label", Type: textType, Arity: published.ParameterAritySingle,
			Default: []resolved.Value{fallback}},
		{Name: "flags", Type: booleanType, Arity: published.ParameterArityList},
		{Name: "payload", Type: jsonType, Arity: published.ParameterAritySingle},
	}, 20)
	if err != nil {
		t.Fatal(err)
	}
	return manifest
}

func requirePublishedScalar(
	t *testing.T, bound published.BoundParameters, name, expected string,
) {
	t.Helper()
	values, found, err := bound.Lookup(name)
	if err != nil || !found || len(values) != 1 {
		t.Fatalf("%s values = %#v, found = %v, err = %v", name, values, found, err)
	}
	actual, scalar, err := values[0].Scalar()
	if err != nil || !scalar || actual != expected {
		t.Fatalf("%s scalar = %q, found = %v, err = %v", name, actual, scalar, err)
	}
}

func requirePublishedEqual(
	t *testing.T, left, right published.BoundParameters,
) {
	t.Helper()
	leftNames, leftErr := left.Names()
	rightNames, rightErr := right.Names()
	if leftErr != nil || rightErr != nil || !reflect.DeepEqual(leftNames, rightNames) {
		t.Fatalf("names differ: %#v %#v, errors: %v %v",
			leftNames, rightNames, leftErr, rightErr)
	}
	for _, name := range leftNames {
		leftValues, _, _ := left.Lookup(name)
		rightValues, _, _ := right.Lookup(name)
		if !reflect.DeepEqual(leftValues, rightValues) {
			t.Fatalf("values differ for %s", name)
		}
	}
	leftPage, leftPresent, leftErr := left.Page()
	rightPage, rightPresent, rightErr := right.Page()
	if leftErr != nil || rightErr != nil || leftPresent != rightPresent {
		t.Fatal("page presence differs")
	}
	if leftPresent {
		leftAfter, _, _ := leftPage.After()
		rightAfter, _, _ := rightPage.After()
		leftSize, _, _ := leftPage.Size()
		rightSize, _, _ := rightPage.Size()
		if leftSize != rightSize || !bytes.Equal([]byte(leftAfter), []byte(rightAfter)) {
			t.Fatal("pages differ")
		}
	}
}
