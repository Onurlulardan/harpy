package httpquery

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type expectedQueryIssue struct {
	path string
	code restqueryv1.IssueCode
}

func requireQueryError(
	t *testing.T,
	document []byte,
	code restqueryv1.ErrorCode,
	issues []expectedQueryIssue,
) api.QueryError {
	t.Helper()
	_, err := NewPOSTDecoder().Decode(document)
	if err == nil {
		t.Fatal("Decode() error = nil")
	}
	var queryError api.QueryError
	if !errors.As(err, &queryError) {
		t.Fatalf("Decode() error type = %T, want api.QueryError", err)
	}
	if queryError.Code() != code {
		t.Fatalf("error code = %q, want %q", queryError.Code(), code)
	}
	actual := queryError.Issues()
	if len(actual) != len(issues) {
		t.Fatalf("issues = %#v, want %d items", actual, len(issues))
	}
	for index, expected := range issues {
		if actual[index].Path != expected.path || actual[index].Code != string(expected.code) {
			t.Fatalf("issue %d = %#v, want path=%q code=%q", index, actual[index], expected.path, expected.code)
		}
	}
	return queryError
}

func decodeSuccess(t *testing.T, document string) api.QueryRequest {
	t.Helper()
	request, err := NewPOSTDecoder().Decode([]byte(document))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	return request
}
