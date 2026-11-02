package postgres

import (
	"strings"
	"testing"
)

func TestIAMProfileRevisionUsesReservedVersions(t *testing.T) {
	want := []string{
		"20260716013200", "20260716013201", "20260716013202",
		"20260716013203", "20260716013204", "20260716013205",
		"20260716013206", "20260716013210", "20260716013211",
		"20260716013212", "20260716013213", "20260716013215",
		"20260716013216", "20260716013217", "20260716013220",
	}
	for index, name := range iamProfileRevisionMigrationNames {
		if !strings.HasPrefix(name, want[index]+"_") {
			t.Fatalf("profile revision migration %q is outside %s", name, want[index])
		}
	}
}
