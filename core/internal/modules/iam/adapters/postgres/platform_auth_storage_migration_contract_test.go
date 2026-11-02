package postgres

import (
	"strings"
	"testing"
)

func assertPlatformAuthMarkers(t *testing.T, source string, markers []string) {
	t.Helper()
	for _, marker := range markers {
		if !strings.Contains(source, marker) {
			t.Fatalf("platform auth migration lacks marker %q", marker)
		}
	}
}
