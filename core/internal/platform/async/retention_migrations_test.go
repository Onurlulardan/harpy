package async

import (
	"strings"
	"testing"
)

func assertRetentionMarkers(t *testing.T, source string, markers []string) {
	t.Helper()
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("retention migration lacks %q", marker)
		}
	}
}

func assertDeleteOrder(t *testing.T, source, child, parent string) {
	t.Helper()
	childIndex, parentIndex := strings.Index(source, child), strings.Index(source, parent)
	if childIndex < 0 || parentIndex < 0 || childIndex >= parentIndex {
		t.Fatalf("retention deletion order is invalid: %q before %q", child, parent)
	}
}
