package v1query

import (
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestV1SortOverridesUseManifestAuthority(t *testing.T) {
	decoder := mustDecoder(t)
	tests := []struct {
		raw       string
		field     string
		direction api.Direction
	}{
		{"sort[field]=status", "status", api.DirectionDescending},
		{"sort[order]=ASC", "created_at", api.DirectionAscending},
		{
			"sort[field]=status&sort[order]=asc",
			"status", api.DirectionAscending,
		},
	}
	for _, test := range tests {
		result, err := decoder.Decode(test.raw)
		if err != nil {
			t.Fatalf("Decode(%q) error = %v", test.raw, err)
		}
		request, _ := result.Request()
		sorts := request.Sort()
		if len(sorts) != 1 || sorts[0].Field() != test.field ||
			sorts[0].Direction() != test.direction {
			t.Fatalf("sort = %#v, want %s/%s", sorts, test.field, test.direction)
		}
	}
}

func TestV1SearchBudgetUsesCanonicalNFCValue(t *testing.T) {
	rawValue := strings.Repeat("e\u0301", 128)
	result, err := mustDecoder(t).Decode("search=" + rawValue)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	search, found := request.Search()
	if !found || search.Term() != strings.Repeat("é", 128) {
		t.Fatalf("canonical search was not NFC-normalized")
	}
}

func TestV1SortUnknownsFailClosed(t *testing.T) {
	decoder := mustDecoder(t)
	tests := []struct {
		raw  string
		path string
		code restqueryv1.IssueCode
	}{
		{"sort[field]=missing", "/sort/field", restqueryv1.IssueFieldUnknown},
		{"sort[order]=sideways", "/sort/order", restqueryv1.IssueValueNotAllowed},
	}
	for _, test := range tests {
		_, err := decoder.Decode(test.raw)
		requireQueryError(
			t, err, restqueryv1.ErrorQueryInvalid, test.path, test.code,
		)
	}
}

func TestV1EmptySearchAndSortLeavesUseSourceDefaults(t *testing.T) {
	result, err := mustDecoder(t).Decode("search=&sort[field]=&sort[order]=")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	request, _ := result.Request()
	if _, found := request.Search(); found {
		t.Fatal("empty V1 search unexpectedly became canonical search")
	}
	sort := request.Sort()[0]
	if sort.Field() != "created_at" || sort.Direction() != api.DirectionDescending {
		t.Fatalf("default sort = %s/%s", sort.Field(), sort.Direction())
	}
}
