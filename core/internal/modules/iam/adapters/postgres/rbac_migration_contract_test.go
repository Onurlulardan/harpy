package postgres

import (
	"strings"
	"testing"
)

func TestIAMRBACUsesOnlyReservedMigrationVersions(t *testing.T) {
	want := []string{
		"20260716013020", "20260716013030", "20260716013040", "20260716013050",
		"20260716013100", "20260716013110", "20260716013120", "20260716013130",
		"20260716013140", "20260716013150",
	}
	for index, name := range iamRBACMigrationNames {
		if !strings.HasPrefix(name, want[index]+"_") {
			t.Fatalf("RBAC migration %q does not use reserved version %s", name, want[index])
		}
	}
}
