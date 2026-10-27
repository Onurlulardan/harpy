package restqueryv1

import (
	"sort"
	"strings"
	"testing"
)

func TestCatalogCodesAreMachineCodesAndDuplicateFree(t *testing.T) {
	t.Parallel()
	var errorCodes []string
	for _, definition := range ErrorCatalog() {
		if !definition.Code.Valid() || definition.HTTPStatus < 400 || definition.HTTPStatus > 599 {
			t.Fatalf("invalid error definition: %#v", definition)
		}
		errorCodes = append(errorCodes, string(definition.Code))
	}
	assertCodeSet(t, errorCodes)

	var issueCodes []string
	for _, code := range IssueCodes() {
		if !code.Valid() {
			t.Fatalf("invalid issue code: %q", code)
		}
		issueCodes = append(issueCodes, string(code))
	}
	assertCodeSet(t, issueCodes)
}

func TestCatalogRejectsMalformedAndUnpublishedCodes(t *testing.T) {
	t.Parallel()
	for _, value := range []string{
		"", "A", "query_invalid", "QUERY-INVALID", "QUERY_INVALID:secret",
		strings.Repeat("A", 65), "QUERY_NOT_PUBLISHED",
	} {
		if ErrorCode(value).Valid() || IssueCode(value).Valid() {
			t.Fatalf("invalid or unpublished code %q accepted", value)
		}
	}
}

func assertCodeSet(t *testing.T, codes []string) {
	t.Helper()
	if !sort.StringsAreSorted(codes) {
		t.Fatalf("catalog is not sorted: %v", codes)
	}
	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if !validMachineCode(code) || seen[code] {
			t.Fatalf("invalid or duplicate machine code: %q", code)
		}
		seen[code] = true
		for _, forbidden := range []string{"PASSWORD", "SECRET", "SQL_TEXT", "IDENTIFIER_VALUE"} {
			if strings.Contains(code, forbidden) {
				t.Fatalf("code carries forbidden data semantics: %q", code)
			}
		}
	}
}
