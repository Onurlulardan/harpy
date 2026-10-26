package httpquery

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

func TestPublishedPOSTRejectsExactTransportFaults(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		name  string
		body  []byte
		issue restqueryv1.IssueCode
	}{
		{"empty", nil, restqueryv1.IssueValueNotCanonical},
		{"malformed", []byte(`{"params":`), restqueryv1.IssueValueNotCanonical},
		{"trailing", []byte(`{} {}`), restqueryv1.IssueValueNotCanonical},
		{"utf8", []byte{'{', '"', 0xff, '"', ':', '1', '}'}, restqueryv1.IssueUTF8Invalid},
		{"size", []byte(strings.Repeat(" ", maximumPublishedPOSTBytes+1)),
			restqueryv1.IssueBodySizeExceeded},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPublishedPOSTDecoder().Decode(test.body, manifest)
			requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryTransportInvalid,
				[]expectedPublishedHTTPIssue{{"", test.issue}})
		})
	}
	nested := strings.Repeat(`{"x":`, maximumJSONMemberNesting+1) + "0" +
		strings.Repeat("}", maximumJSONMemberNesting+1)
	_, err := NewPublishedPOSTDecoder().Decode([]byte(nested), manifest)
	requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryTransportInvalid,
		[]expectedPublishedHTTPIssue{{"", restqueryv1.IssueValueNotCanonical}})
}

func TestPublishedPOSTRejectsDuplicatesAtEveryDepth(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		body string
		path string
	}{
		{`{"params":{"count":"1"},"params":{"count":"2"}}`, "/params"},
		{`{"params":{"count":"1","count":"2"}}`, "/params/count"},
		{`{"params":{"count":"1","payload":{"a":1,"a":2}}}`, "/params/payload/a"},
	}
	for _, test := range tests {
		_, err := NewPublishedPOSTDecoder().Decode([]byte(test.body), manifest)
		requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryDuplicateMember,
			[]expectedPublishedHTTPIssue{
				{test.path, restqueryv1.IssueMemberDuplicate},
			})
	}
}

func TestPublishedPOSTEnforcesCursorAndBodyBoundaries(t *testing.T) {
	manifest := publishedTestManifest(t)
	cursor := strings.Repeat("x", 16*1024)
	body := `{"params":{"count":"1"},"page":{"size":10,"after":"` + cursor + `"}}`
	if _, err := NewPublishedPOSTDecoder().Decode([]byte(body), manifest); err != nil {
		t.Fatalf("cursor boundary rejected: %v", err)
	}
	cursor += "x"
	body = `{"params":{"count":"1"},"page":{"size":10,"after":"` + cursor + `"}}`
	_, err := NewPublishedPOSTDecoder().Decode([]byte(body), manifest)
	requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
		[]expectedPublishedHTTPIssue{
			{"/page/after", restqueryv1.IssueCursorSizeExceeded},
		})
	prefix := `{"params":{"count":"1","label":"`
	suffix := `"}}`
	exact := prefix + strings.Repeat("x", maximumPublishedPOSTBytes-len(prefix)-len(suffix)) + suffix
	if len(exact) != maximumPublishedPOSTBytes {
		t.Fatal("body is not exact boundary")
	}
	if _, err := NewPublishedPOSTDecoder().Decode([]byte(exact), manifest); err != nil {
		t.Fatalf("body boundary rejected: %v", err)
	}
}

func TestPublishedDecodersRejectInvalidManifestBeforeInput(t *testing.T) {
	if _, err := NewPublishedGETDecoder().Decode("", published.ParameterManifest{}); err == nil {
		t.Fatal("GET accepted zero manifest")
	}
	if _, err := NewPublishedPOSTDecoder().Decode(nil, published.ParameterManifest{}); err == nil {
		t.Fatal("POST accepted zero manifest")
	}
}
