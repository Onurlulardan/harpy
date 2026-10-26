package httpquery

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestGETDecoderRejectsExactTransportFaults(t *testing.T) {
	tests := []struct {
		name  string
		raw   string
		issue restqueryv1.IssueCode
	}{
		{"size", strings.Repeat("x", maximumGETQueryBytes+1), restqueryv1.IssueQuerySizeExceeded},
		{"percent", "count=%ZZ", restqueryv1.IssuePercentEncodingInvalid},
		{"raw plus", "search[term]=a+b", restqueryv1.IssueRawPlusForbidden},
		{"raw space", "search[term]=a b", restqueryv1.IssuePercentEncodingInvalid},
		{"nul", "search[term]=a%00b", restqueryv1.IssueNULForbidden},
		{"decoded utf8", "search[term]=%FF", restqueryv1.IssueUTF8Invalid},
		{"raw utf8", "search[term]=" + string([]byte{0xff}), restqueryv1.IssueUTF8Invalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requireGETError(t, test.raw, restqueryv1.ErrorQueryTransportInvalid,
				[]expectedGETIssue{{"", test.issue}})
		})
	}
}

func TestGETDecoderAcceptsExactRawSizeBoundary(t *testing.T) {
	prefix := "filter[0][field]=x&filter[0][op]=eq&filter[0][value]="
	rawQuery := prefix + strings.Repeat("a", maximumGETQueryBytes-len(prefix))
	if len(rawQuery) != maximumGETQueryBytes {
		t.Fatal("test query is not the exact boundary")
	}
	decodeGETSuccess(t, rawQuery)
}

func TestGETDecoderRequiresOneRawEqualsPerLeaf(t *testing.T) {
	for _, rawQuery := range []string{"count", "count=none=extra", "count=none&"} {
		requireGETError(t, rawQuery, restqueryv1.ErrorQueryInvalid,
			[]expectedGETIssue{{"", restqueryv1.IssueValueNotCanonical}})
	}
	request := decodeGETSuccess(
		t, "filter[0][field]=x&filter[0][op]=eq&filter[0][value]=a%3Db",
	)
	where, _ := request.Where()
	predicate, _ := where.Children()[0].Predicate()
	if value := getTextValue(t, predicate.Values()[0]); value != "a=b" {
		t.Fatalf("encoded equals changed: %q", value)
	}
}

func TestGETDecoderAllowsOnlyEncodedSpaceAndPlus(t *testing.T) {
	request := decodeGETSuccess(t, "search[term]=a%20b")
	search, present := request.Search()
	if !present || search.Term() != "a b" {
		t.Fatal("encoded space changed")
	}
	request = decodeGETSuccess(t,
		"filter[0][field]=x&filter[0][op]=eq&filter[0][value]=a%2Bb",
	)
	where, _ := request.Where()
	predicate, _ := where.Children()[0].Predicate()
	if value := getTextValue(t, predicate.Values()[0]); value != "a+b" {
		t.Fatalf("encoded plus changed: %q", value)
	}
}
