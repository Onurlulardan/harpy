package api

import (
	"fmt"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestQueryErrorSortsDeduplicatesThenCapsIssues(t *testing.T) {
	input := make([]runtimev1.Issue, 0, 27)
	for index := 24; index >= 0; index-- {
		input = append(input, runtimev1.Issue{
			Path: fmt.Sprintf("/field/%02d", index), Code: "FIELD_UNKNOWN",
		})
	}
	input = append(input, input[3], input[3])
	queryError, err := NewQueryError(restqueryv1.ErrorQueryInvalid, input)
	if err != nil {
		t.Fatal(err)
	}
	issues := queryError.Issues()
	if len(issues) != maximumQueryIssues {
		t.Fatalf("issue count = %d, want %d", len(issues), maximumQueryIssues)
	}
	for index, issue := range issues {
		want := fmt.Sprintf("/field/%02d", index)
		if issue.Path != want {
			t.Fatalf("issue %d path = %q, want %q", index, issue.Path, want)
		}
	}
}

func TestQueryErrorSortsByPathThenCodeAndRemovesExactDuplicates(t *testing.T) {
	input := []runtimev1.Issue{
		{Path: "/b", Code: "TYPE_MISMATCH"},
		{Path: "/a", Code: "TYPE_MISMATCH"},
		{Path: "/a", Code: "FIELD_UNKNOWN"},
		{Path: "/a", Code: "FIELD_UNKNOWN"},
	}
	queryError, err := NewQueryError(restqueryv1.ErrorQueryInvalid, input)
	if err != nil {
		t.Fatal(err)
	}
	want := []runtimev1.Issue{
		{Path: "/a", Code: "FIELD_UNKNOWN"},
		{Path: "/a", Code: "TYPE_MISMATCH"},
		{Path: "/b", Code: "TYPE_MISMATCH"},
	}
	if !equalIssues(queryError.Issues(), want) {
		t.Fatalf("issues = %#v, want %#v", queryError.Issues(), want)
	}
}

func TestQueryErrorOwnsInputAndOutputSlices(t *testing.T) {
	input := []runtimev1.Issue{{Path: "/field", Code: "FIELD_UNKNOWN"}}
	queryError, err := NewQueryError(restqueryv1.ErrorQueryInvalid, input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = runtimev1.Issue{Path: "/changed", Code: "TYPE_MISMATCH"}
	output := queryError.Issues()
	output[0] = runtimev1.Issue{}
	stored := queryError.Issues()
	if stored[0].Path != "/field" || stored[0].Code != "FIELD_UNKNOWN" {
		t.Fatalf("stored issue changed: %#v", stored[0])
	}
}

func equalIssues(left, right []runtimev1.Issue) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
