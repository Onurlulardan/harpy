package httpquery

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPublishedPOSTRequiresExactRootAndParameterObject(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		name   string
		body   string
		issues []expectedPublishedHTTPIssue
	}{
		{"root type", `[]`, []expectedPublishedHTTPIssue{
			{"", restqueryv1.IssueTypeMismatch},
		}},
		{"missing params", `{}`, []expectedPublishedHTTPIssue{
			{"/params", restqueryv1.IssueParameterRequired},
		}},
		{"params type", `{"params":[]}`, []expectedPublishedHTTPIssue{
			{"/params", restqueryv1.IssueTypeMismatch},
		}},
		{"unknown root", `{"params":{"count":"1"},"other":1}`,
			[]expectedPublishedHTTPIssue{
				{"/other", restqueryv1.IssueParameterUnknown},
			}},
		{"unknown parameter", `{"params":{"count":"1","other":1}}`,
			[]expectedPublishedHTTPIssue{
				{"/params/other", restqueryv1.IssueParameterUnknown},
			}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPublishedPOSTDecoder().Decode([]byte(test.body), manifest)
			requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
				test.issues)
		})
	}
}

func TestPublishedPOSTUsesManifestArityAndExactTypedValues(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		name  string
		body  string
		path  string
		issue restqueryv1.IssueCode
	}{
		{"numeric integer", `{"params":{"count":1}}`, "/params/count",
			restqueryv1.IssueTypeMismatch},
		{"noncanonical integer", `{"params":{"count":"01"}}`, "/params/count",
			restqueryv1.IssueValueNotCanonical},
		{"integer array", `{"params":{"count":["1"]}}`, "/params/count",
			restqueryv1.IssueTypeMismatch},
		{"list scalar", `{"params":{"count":"1","flags":true}}`, "/params/flags",
			restqueryv1.IssueParameterArityMismatch},
		{"empty list", `{"params":{"count":"1","flags":[]}}`, "/params/flags",
			restqueryv1.IssueParameterArityMismatch},
		{"null scalar", `{"params":{"count":null}}`, "/params/count",
			restqueryv1.IssueValueNotAllowed},
		{"null list item", `{"params":{"count":"1","flags":[true,null]}}`,
			"/params/flags/1", restqueryv1.IssueValueNotAllowed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPublishedPOSTDecoder().Decode([]byte(test.body), manifest)
			requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
				[]expectedPublishedHTTPIssue{{test.path, test.issue}})
		})
	}
	acceptedItems := strings.Repeat("true,", 99) + "true"
	acceptedBody := `{"params":{"count":"1","flags":[` + acceptedItems + `]}}`
	if _, err := NewPublishedPOSTDecoder().Decode([]byte(acceptedBody), manifest); err != nil {
		t.Fatalf("100 list values rejected: %v", err)
	}
	items := strings.Repeat("true,", 100) + "true"
	body := `{"params":{"count":"1","flags":[` + items + `]}}`
	_, err := NewPublishedPOSTDecoder().Decode([]byte(body), manifest)
	requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
		[]expectedPublishedHTTPIssue{
			{"/params/flags", restqueryv1.IssueParameterArityMismatch},
		})
}

func TestPublishedPOSTRequiresExactPageObject(t *testing.T) {
	manifest := publishedTestManifest(t)
	tests := []struct {
		name   string
		body   string
		issues []expectedPublishedHTTPIssue
	}{
		{"page type", `{"params":{"count":"1"},"page":null}`,
			[]expectedPublishedHTTPIssue{{"/page", restqueryv1.IssueTypeMismatch}}},
		{"missing after", `{"params":{"count":"1"},"page":{"size":10}}`,
			[]expectedPublishedHTTPIssue{{"/page/after", restqueryv1.IssueParameterRequired}}},
		{"missing size", `{"params":{"count":"1"},"page":{"after":null}}`,
			[]expectedPublishedHTTPIssue{{"/page/size", restqueryv1.IssueParameterRequired}}},
		{"unknown", `{"params":{"count":"1"},"page":{"size":10,"after":null,"x":1}}`,
			[]expectedPublishedHTTPIssue{{"/page/x", restqueryv1.IssueParameterUnknown}}},
		{"size string", `{"params":{"count":"1"},"page":{"size":"10","after":null}}`,
			[]expectedPublishedHTTPIssue{{"/page/size", restqueryv1.IssueTypeMismatch}}},
		{"size zero", `{"params":{"count":"1"},"page":{"size":0,"after":null}}`,
			[]expectedPublishedHTTPIssue{{"/page/size", restqueryv1.IssueValueOutOfRange}}},
		{"manifest max", `{"params":{"count":"1"},"page":{"size":21,"after":null}}`,
			[]expectedPublishedHTTPIssue{{"/page/size", restqueryv1.IssuePageSizeExceeded}}},
		{"empty after", `{"params":{"count":"1"},"page":{"size":10,"after":""}}`,
			[]expectedPublishedHTTPIssue{{"/page/after", restqueryv1.IssueValueNotAllowed}}},
		{"after type", `{"params":{"count":"1"},"page":{"size":10,"after":1}}`,
			[]expectedPublishedHTTPIssue{{"/page/after", restqueryv1.IssueTypeMismatch}}},
		{"after NUL", `{"params":{"count":"1"},"page":{"size":10,"after":"\u0000"}}`,
			[]expectedPublishedHTTPIssue{{"/page/after", restqueryv1.IssueNULForbidden}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewPublishedPOSTDecoder().Decode([]byte(test.body), manifest)
			requirePublishedHTTPError(t, err, restqueryv1.ErrorQueryParameterInvalid,
				test.issues)
		})
	}
	bound, err := NewPublishedPOSTDecoder().Decode(
		[]byte(`{"params":{"count":"1"},"page":{"size":10,"after":null}}`), manifest,
	)
	if err != nil {
		t.Fatalf("null initial cursor rejected: %v", err)
	}
	page, present, _ := bound.Page()
	_, hasAfter, _ := page.After()
	if !present || hasAfter {
		t.Fatal("null cursor did not normalize to absent")
	}
}
