package httpquery

import (
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPublishedGETReportsUnknownMissingAndArityErrors(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		name   string
		raw    string
		issues []expectedPublishedHTTPIssue
	}{
		{"missing", "", []expectedPublishedHTTPIssue{
			{"/param/count", restqueryv1.IssueParameterRequired},
		}},
		{"unknown root", "param[count]=1&other=x",
			[]expectedPublishedHTTPIssue{
				{"/other", restqueryv1.IssueParameterUnknown},
			}},
		{"unknown parameter", "param[count]=1&param[nope]=x",
			[]expectedPublishedHTTPIssue{
				{"/param/nope", restqueryv1.IssueParameterUnknown},
			}},
		{"single encoded as list", "param[count][0]=1",
			[]expectedPublishedHTTPIssue{
				{"/param/count", restqueryv1.IssueParameterArityMismatch},
			}},
		{"list encoded as single", "param[count]=1&param[flags]=true",
			[]expectedPublishedHTTPIssue{
				{"/param/flags", restqueryv1.IssueParameterArityMismatch},
			}},
		{"mixed list shape", "param[count]=1&param[flags]=true&param[flags][0]=false",
			[]expectedPublishedHTTPIssue{
				{"/param/flags", restqueryv1.IssueParameterArityMismatch},
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPublishedGETDecoder().Decode(test.raw, manifest)
			requirePublishedHTTPError(
				t, err, restqueryv1.ErrorQueryParameterInvalid, test.issues,
			)
		})
	}
}

func TestPublishedGETEnforcesCanonicalListIndices(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		name  string
		raw   string
		path  string
		issue restqueryv1.IssueCode
	}{
		{"leading zero", "param[count]=1&param[flags][00]=true",
			"/param/flags/00", restqueryv1.IssueIndexInvalid},
		{"nondigit", "param[count]=1&param[flags][x]=true",
			"/param/flags/x", restqueryv1.IssueIndexInvalid},
		{"sparse", "param[count]=1&param[flags][0]=true&param[flags][2]=false",
			"/param/flags/2", restqueryv1.IssueIndexNonContiguous},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPublishedGETDecoder().Decode(test.raw, manifest)
			requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
				[]expectedPublishedHTTPIssue{{test.path, test.issue}})
		})
	}
	query := publishedGETListQuery(100)
	if _, err := NewPublishedGETDecoder().Decode(query, manifest); err != nil {
		t.Fatalf("100 list values rejected: %v", err)
	}
	_, err := NewPublishedGETDecoder().Decode(publishedGETListQuery(101), manifest)
	requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
		[]expectedPublishedHTTPIssue{
			{"/param/flags", restqueryv1.IssueParameterArityMismatch},
		})
}

func publishedGETListQuery(count int) string {
	var query strings.Builder
	query.WriteString("param[count]=1")
	for index := 0; index < count; index++ {
		fmt.Fprintf(&query, "&param[flags][%d]=true", index)
	}
	return query.String()
}

func TestPublishedGETEnforcesPageContract(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		raw   string
		path  string
		issue restqueryv1.IssueCode
	}{
		{"param[count]=1&page[size]=01", "/page/size", restqueryv1.IssueValueNotCanonical},
		{"param[count]=1&page[size]=0", "/page/size", restqueryv1.IssueValueOutOfRange},
		{"param[count]=1&page[size]=21", "/page/size", restqueryv1.IssuePageSizeExceeded},
		{"param[count]=1&page[size]=501", "/page/size", restqueryv1.IssuePageSizeExceeded},
		{"param[count]=1&page[after]=", "/page/after", restqueryv1.IssueValueNotAllowed},
		{"param[count]=1&page[before]=x", "/page/before", restqueryv1.IssueParameterUnknown},
	}
	for _, test := range tests {
		_, err := NewPublishedGETDecoder().Decode(test.raw, manifest)
		requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
			[]expectedPublishedHTTPIssue{{test.path, test.issue}})
	}
	bound, err := NewPublishedGETDecoder().Decode(
		"param[count]=1&page[after]=opaque", manifest,
	)
	if err != nil {
		t.Fatalf("after-only page rejected: %v", err)
	}
	page, present, err := bound.Page()
	if err != nil || !present {
		t.Fatalf("page present = %v, err = %v", present, err)
	}
	if _, hasSize, _ := page.Size(); hasSize {
		t.Fatal("after-only page gained a size")
	}
}

func TestPublishedGETPercentDecodesExactlyOnce(t *testing.T) {
	bound, err := NewPublishedGETDecoder().Decode(
		"param[count]=1&param[label]=%2520", publishedTestManifest(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	requirePublishedScalar(t, bound, "label", "%20")
}
