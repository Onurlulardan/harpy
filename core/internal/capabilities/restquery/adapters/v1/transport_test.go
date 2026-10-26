package v1query

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestV1WireDecodesPlusAndPercentExactlyOnce(t *testing.T) {
	decoder := mustDecoder(t)
	result, err := decoder.Decode("search=Acme+%2520+Corp")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	search, _ := request.Search()
	if search.Term() != "Acme %20 Corp" {
		t.Fatalf("search term = %q, want exact one-pass decode", search.Term())
	}
}

func TestV1WireTransportFailuresUseStableErrors(t *testing.T) {
	decoder := mustDecoder(t)
	tests := []struct {
		name string
		raw  string
		code restqueryv1.IssueCode
	}{
		{"percent", "search=%zz", restqueryv1.IssuePercentEncodingInvalid},
		{"UTF-8", "search=%ff", restqueryv1.IssueUTF8Invalid},
		{"NUL", "search=%00", restqueryv1.IssueNULForbidden},
		{
			"size", "search=" + strings.Repeat("a", maximumV1QueryBytes),
			restqueryv1.IssueQuerySizeExceeded,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decoder.Decode(test.raw)
			requireQueryError(
				t, err, restqueryv1.ErrorQueryTransportInvalid, "", test.code,
			)
		})
	}
}

func TestV1UnknownMembersAndFieldsUseExactRFC6901Paths(t *testing.T) {
	decoder := mustDecoder(t)
	tests := []struct {
		raw  string
		path string
		code restqueryv1.IssueCode
	}{
		{"unknown[a/b~c]=x", "/unknown/a~1b~0c", restqueryv1.IssueMemberUnknown},
		{
			"filters[a/b~c]=x", "/filters/a~1b~0c",
			restqueryv1.IssueFieldUnknown,
		},
	}
	for _, test := range tests {
		_, err := decoder.Decode(test.raw)
		requireQueryError(
			t, err, restqueryv1.ErrorQueryInvalid, test.path, test.code,
		)
	}
}

func TestV1RepeatedScalarKeepsSourceFirstValue(t *testing.T) {
	result, err := mustDecoder(t).Decode("search=first&search=second")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	search, _ := request.Search()
	if search.Term() != "first" {
		t.Fatalf("search term = %q, want first source value", search.Term())
	}
}

func TestV1SearchBudgetIsNotTruncated(t *testing.T) {
	_, err := mustDecoder(t).Decode("search=" + strings.Repeat("a", 129))
	requireQueryError(
		t, err, restqueryv1.ErrorQueryBudgetExceeded,
		"/search", restqueryv1.IssueSearchSizeExceeded,
	)
}
