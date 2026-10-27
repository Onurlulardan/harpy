package httpquery

import (
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestGETDecoderRejectsNonCanonicalShapes(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		path  string
		issue restqueryv1.IssueCode
	}{
		{"include", "include[0]=x", "/include", restqueryv1.IssueMemberForbidden},
		{"aggregate", "aggregate=x", "/aggregate", restqueryv1.IssueMemberForbidden},
		{"nested boolean", "filter[0][and]=x", "/filter/0/and", restqueryv1.IssueMemberForbidden},
		{"unknown root", "where=x", "/where", restqueryv1.IssueMemberUnknown},
		{"unindexed select", "select=x", "/select", restqueryv1.IssueIndexInvalid},
		{"leading zero", "select[01]=x", "/select/01", restqueryv1.IssueIndexInvalid},
		{"negative", "sort[-1][field]=x", "/sort/-1", restqueryv1.IssueIndexInvalid},
		{"sparse", "select[1]=x", "/select/1", restqueryv1.IssueIndexNonContiguous},
		{"search shorthand", "search=x", "/search", restqueryv1.IssueValueNotCanonical},
		{"count object", "count[x]=none", "/count/x", restqueryv1.IssueValueNotCanonical},
		{"unknown filter member", "filter[0][x]=y", "/filter/0/x", restqueryv1.IssueMemberUnknown},
		{"unindexed values", "filter[0][values]=x", "/filter/0/values", restqueryv1.IssueIndexInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireGETError(t, test.raw, restqueryv1.ErrorQueryInvalid,
				[]expectedGETIssue{{test.path, test.issue}})
		})
	}
}

func TestGETDecoderRejectsSparseOperandIndices(t *testing.T) {
	rawQuery := "filter[0][field]=x&filter[0][op]=in&filter[0][values][1]=a"
	requireGETError(t, rawQuery, restqueryv1.ErrorQueryInvalid,
		[]expectedGETIssue{{"/filter/0/values/1", restqueryv1.IssueIndexNonContiguous}})
}

func TestGETDecoderRejectsDuplicateSelectAndSortFields(t *testing.T) {
	tests := []struct {
		raw  string
		path string
	}{
		{"select[0]=x&select[1]=x", "/select/1"},
		{"sort[0][field]=x&sort[0][direction]=asc&" +
			"sort[1][field]=x&sort[1][direction]=desc", "/sort/1/field"},
	}
	for _, test := range tests {
		requireGETError(t, test.raw, restqueryv1.ErrorQueryInvalid,
			[]expectedGETIssue{{test.path, restqueryv1.IssueFieldDuplicate}})
	}
}

func TestGETDecoderEnforcesRequiredMembersAndArity(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		issues []expectedGETIssue
	}{
		{"missing value", "filter[0][field]=x&filter[0][op]=eq",
			[]expectedGETIssue{{"/filter/0/value", restqueryv1.IssueMemberRequired}}},
		{"zero with value", "filter[0][field]=x&filter[0][op]=is_null&filter[0][value]=x",
			[]expectedGETIssue{{"/filter/0/value", restqueryv1.IssueArityMismatch}}},
		{"list with scalar", "filter[0][field]=x&filter[0][op]=in&filter[0][value]=x",
			[]expectedGETIssue{
				{"/filter/0/value", restqueryv1.IssueArityMismatch},
				{"/filter/0/values", restqueryv1.IssueMemberRequired},
			}},
		{"pair count", "filter[0][field]=x&filter[0][op]=between&filter[0][values][0]=x",
			[]expectedGETIssue{{"/filter/0/values", restqueryv1.IssueArityMismatch}}},
		{"unknown operator", "filter[0][field]=x&filter[0][op]=wat",
			[]expectedGETIssue{{"/filter/0/op", restqueryv1.IssueOperatorUnknown}}},
		{"missing sort direction", "sort[0][field]=x",
			[]expectedGETIssue{{"/sort/0/direction", restqueryv1.IssueMemberRequired}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireGETError(t, test.raw, restqueryv1.ErrorQueryInvalid, test.issues)
		})
	}
}

func TestGETDecoderEnforcesScalarGrammar(t *testing.T) {
	tests := []struct {
		raw   string
		path  string
		issue restqueryv1.IssueCode
	}{
		{"page[size]=01", "/page/size", restqueryv1.IssueValueNotCanonical},
		{"page[size]=-1", "/page/size", restqueryv1.IssueValueNotCanonical},
		{"page[size]=0", "/page/size", restqueryv1.IssueValueOutOfRange},
		{"page[after]=", "/page/after", restqueryv1.IssueValueNotAllowed},
		{"count=all", "/count", restqueryv1.IssueValueNotAllowed},
		{"search[term]=%20", "/search/term", restqueryv1.IssueValueNotAllowed},
	}
	for _, test := range tests {
		requireGETError(t, test.raw, restqueryv1.ErrorQueryInvalid,
			[]expectedGETIssue{{test.path, test.issue}})
	}
}

func TestGETDecoderErrorsDoNotReflectRejectedValues(t *testing.T) {
	secret := "private-secret-value"
	err := requireGETError(t, "count="+secret, restqueryv1.ErrorQueryInvalid,
		[]expectedGETIssue{{"/count", restqueryv1.IssueValueNotAllowed}})
	for _, formatted := range []string{fmt.Sprint(err), fmt.Sprintf("%#v", err)} {
		if strings.Contains(formatted, secret) {
			t.Fatalf("GET error leaked rejected value: %q", formatted)
		}
	}
}
