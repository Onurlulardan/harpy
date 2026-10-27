package systemresourcequery

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestEnsureGETSelectAppendsMandatoryFields(t *testing.T) {
	request, err := DecodeGET("select[0]=display_name")
	if err != nil {
		t.Fatal(err)
	}
	rewritten, err := EnsureGETSelect(request, []string{"id", "revision"})
	if err != nil {
		t.Fatal(err)
	}
	value, err := rewritten.Value()
	if err != nil {
		t.Fatal(err)
	}
	assertSelect(t, value.Select(), []string{"display_name", "id", "revision"})
}

func TestEnsureGETSelectPreservesDefaultAndDoesNotDuplicate(t *testing.T) {
	defaultRequest, err := DecodeGET("")
	if err != nil {
		t.Fatal(err)
	}
	unchanged, err := EnsureGETSelect(defaultRequest, []string{"id", "revision"})
	if err != nil {
		t.Fatal(err)
	}
	defaultValue, _ := unchanged.Value()
	assertSelect(t, defaultValue.Select(), []string{})

	explicit, err := DecodeGET("select[0]=id&select[1]=display_name")
	if err != nil {
		t.Fatal(err)
	}
	rewritten, err := EnsureGETSelect(explicit, []string{"id", "revision"})
	if err != nil {
		t.Fatal(err)
	}
	value, _ := rewritten.Value()
	assertSelect(t, value.Select(), []string{"id", "display_name", "revision"})
}

func TestEnsureGETSelectFailsWhenMandatoryFieldsExceedBudget(t *testing.T) {
	request, err := DecodeGET(selectQuery(49))
	if err != nil {
		t.Fatal(err)
	}
	_, err = EnsureGETSelect(request, []string{"id", "revision"})
	var failure Failure
	if !errors.As(err, &failure) ||
		failure.Code() != string(restqueryv1.ErrorQueryBudgetExceeded) ||
		failure.HTTPStatus() != 422 ||
		len(failure.Issues()) != 1 ||
		failure.Issues()[0].Path != "/select" ||
		failure.Issues()[0].Code != string(restqueryv1.IssueSelectCountExceeded) {
		t.Fatalf("unexpected select budget failure: %#v", err)
	}
}

func assertSelect(t *testing.T, actual, expected []string) {
	t.Helper()
	if strings.Join(actual, ",") != strings.Join(expected, ",") {
		t.Fatalf("select mismatch: got %v want %v", actual, expected)
	}
}

func selectQuery(count int) string {
	values := make([]string, count)
	for index := range values {
		values[index] = "select[" + strconv.Itoa(index) + "]=field_" +
			strconv.Itoa(index)
	}
	return strings.Join(values, "&")
}
