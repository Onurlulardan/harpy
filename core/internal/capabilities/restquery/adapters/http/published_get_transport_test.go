package httpquery

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPublishedGETRejectsExactTransportFaults(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		name  string
		raw   string
		issue restqueryv1.IssueCode
	}{
		{"size", strings.Repeat("x", maximumGETQueryBytes+1),
			restqueryv1.IssueQuerySizeExceeded},
		{"percent", "param[count]=%ZZ", restqueryv1.IssuePercentEncodingInvalid},
		{"raw plus", "param[count]=1+2", restqueryv1.IssueRawPlusForbidden},
		{"raw space", "param[label]=a b", restqueryv1.IssuePercentEncodingInvalid},
		{"nul", "param[label]=a%00b", restqueryv1.IssueNULForbidden},
		{"decoded utf8", "param[label]=%FF", restqueryv1.IssueUTF8Invalid},
		{"raw utf8", "param[label]=" + string([]byte{0xff}), restqueryv1.IssueUTF8Invalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPublishedGETDecoder().Decode(test.raw, manifest)
			requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryTransportInvalid,
				[]expectedPublishedHTTPIssue{{"", test.issue}})
		})
	}
}

func TestPublishedGETRejectsMalformedLeavesAndDecodedDuplicates(t *testing.T) {
	manifest := publishedTestManifest(t)
	for _, raw := range []string{"param[count]", "param[count]=1=2", "param[count]=1&"} {
		_, err := NewPublishedGETDecoder().Decode(raw, manifest)
		requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
			[]expectedPublishedHTTPIssue{{"", restqueryv1.IssueValueNotCanonical}})
	}
	_, err := NewPublishedGETDecoder().Decode(
		"param[count]=1&param%5Bcount%5D=2", manifest,
	)
	requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryDuplicateMember,
		[]expectedPublishedHTTPIssue{
			{"/param/count", restqueryv1.IssueMemberDuplicate},
		})
}

func TestPublishedGETRequiresStructuralNFC(t *testing.T) {
	_, err := NewPublishedGETDecoder().Decode(
		"param[e%CC%81]=x", publishedTestManifest(t),
	)
	requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
		[]expectedPublishedHTTPIssue{
			{"/param/\u00e9", restqueryv1.IssueNFCRequired},
		})
}

func TestPublishedGETAllowsEncodedSpaceAndPlus(t *testing.T) {
	bound, err := NewPublishedGETDecoder().Decode(
		"param[count]=1&param[label]=a%20b%2Bc", publishedTestManifest(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	requirePublishedScalar(t, bound, "label", "a b+c")
}

func TestPublishedGETAcceptsExactRawSizeBoundary(t *testing.T) {
	prefix := "param[count]=1&param[label]="
	raw := prefix + strings.Repeat("a", maximumGETQueryBytes-len(prefix))
	if len(raw) != maximumGETQueryBytes {
		t.Fatal("query is not exact boundary")
	}
	if _, err := NewPublishedGETDecoder().Decode(raw, publishedTestManifest(t)); err != nil {
		t.Fatalf("boundary rejected: %v", err)
	}
}
