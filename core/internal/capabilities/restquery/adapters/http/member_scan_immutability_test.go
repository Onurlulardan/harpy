package httpquery

import (
	"slices"
	"sort"
	"testing"
)

func TestMemberScanReturnsIndependentDuplicatePathCopies(t *testing.T) {
	scan, err := scanJSONMembers([]byte(
		`{"first":1,"first":2,"nested":{"second":1,"second":2}}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"/first", "/nested/second"}
	first := scan.duplicatePaths()
	if len(first) != len(want) {
		t.Fatalf("duplicate path count = %d", len(first))
	}
	first[0] = "/mutated"
	first = append(first, "/injected")

	copied := scan
	second := copied.duplicatePaths()
	sort.Strings(second)
	if !slices.Equal(second, want) {
		t.Fatalf("stored duplicate paths were mutated: %v", second)
	}
}
