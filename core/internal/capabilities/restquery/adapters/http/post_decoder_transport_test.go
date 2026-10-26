package httpquery

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderRejectsTransportFailures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		document []byte
		issue    restqueryv1.IssueCode
	}{
		{"empty", nil, restqueryv1.IssueValueNotCanonical},
		{"malformed", []byte(`{"where":`), restqueryv1.IssueValueNotCanonical},
		{"trailing", []byte(`{} {}`), restqueryv1.IssueValueNotCanonical},
		{"invalid UTF-8", []byte{'{', '"', 0xff, '"', ':', '1', '}'}, restqueryv1.IssueUTF8Invalid},
		{"too large", []byte(strings.Repeat(" ", maximumPOSTDocumentBytes+1)), restqueryv1.IssueBodySizeExceeded},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			requireQueryError(t, test.document, restqueryv1.ErrorQueryTransportInvalid,
				[]expectedQueryIssue{{"", test.issue}})
		})
	}
}

func TestPOSTDecoderRejectsExcessiveJSONNesting(t *testing.T) {
	t.Parallel()
	document := strings.Repeat(`{"x":`, maximumJSONMemberNesting+1) + `0` +
		strings.Repeat(`}`, maximumJSONMemberNesting+1)
	requireQueryError(t, []byte(document), restqueryv1.ErrorQueryTransportInvalid,
		[]expectedQueryIssue{{"", restqueryv1.IssueValueNotCanonical}})
}

func TestPOSTDecoderRequiresRootObject(t *testing.T) {
	t.Parallel()
	requireQueryError(t, []byte(`[]`), restqueryv1.ErrorQueryInvalid,
		[]expectedQueryIssue{{"", restqueryv1.IssueTypeMismatch}})
}

func TestPOSTDecoderAcceptsExactBodySizeBoundary(t *testing.T) {
	t.Parallel()
	document := []byte(strings.Repeat(" ", maximumPOSTDocumentBytes-2) + `{}`)
	request, err := NewPOSTDecoder().Decode(document)
	if err != nil {
		t.Fatalf("Decode() boundary error = %v", err)
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("boundary request is invalid: %v", err)
	}
}
