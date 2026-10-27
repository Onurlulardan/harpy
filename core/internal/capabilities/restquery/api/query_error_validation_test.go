package api

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestQueryErrorRejectsInvalidOrUnpublishedOuterCode(t *testing.T) {
	for _, code := range []restqueryv1.ErrorCode{
		"", "query_invalid", "QUERY_NOT_PUBLISHED",
	} {
		if _, err := NewQueryError(code, nil); err == nil {
			t.Fatalf("outer code %q accepted", code)
		}
	}
	queryError, err := NewQueryError(restqueryv1.ErrorQueryFieldForbidden, nil)
	if err != nil {
		t.Fatal(err)
	}
	status, found := queryError.HTTPStatus()
	if queryError.Code() != restqueryv1.ErrorQueryFieldForbidden ||
		!found || status != 403 || queryError.Issues() == nil {
		t.Fatalf("unexpected query error: %#v", queryError)
	}
	if status, found := (QueryError{}).HTTPStatus(); found || status != 0 {
		t.Fatalf("zero query error status = %d, found = %v", status, found)
	}
}

func TestQueryErrorValidatesEveryIssueBeforeCap(t *testing.T) {
	issues := make([]runtimev1.Issue, 25)
	for index := range issues {
		issues[index] = runtimev1.Issue{Path: "/field", Code: "FIELD_UNKNOWN"}
	}
	issues[24].Path = "not-a-pointer"
	if _, err := NewQueryError(restqueryv1.ErrorQueryInvalid, issues); err == nil {
		t.Fatal("invalid issue beyond cap was ignored")
	}
	issues[24] = runtimev1.Issue{Path: "/field", Code: "lowercase"}
	if _, err := NewQueryError(restqueryv1.ErrorQueryInvalid, issues); err == nil {
		t.Fatal("invalid issue machine code was accepted")
	}
}

func TestNewQueryIssueUsesClosedCatalogAndRFC6901Pointer(t *testing.T) {
	issue, err := NewQueryIssue("/where/and/0", restqueryv1.IssueFieldUnknown)
	if err != nil {
		t.Fatal(err)
	}
	if issue.Path != "/where/and/0" || issue.Code != "FIELD_UNKNOWN" {
		t.Fatalf("unexpected issue: %#v", issue)
	}
	if _, err := NewQueryIssue("where", restqueryv1.IssueFieldUnknown); err == nil {
		t.Fatal("invalid JSON pointer accepted")
	}
	if _, err := NewQueryIssue("/where", "ISSUE_NOT_PUBLISHED"); err == nil {
		t.Fatal("unpublished issue code accepted")
	}
}

func TestQueryErrorRejectsUnpublishedRuntimeMachineCode(t *testing.T) {
	issue := runtimev1.Issue{Path: "/where", Code: "EXTENSION_VALID"}
	if _, err := NewQueryError(restqueryv1.ErrorQueryInvalid, []runtimev1.Issue{issue}); err == nil {
		t.Fatal("unpublished runtime issue code accepted")
	}
}
