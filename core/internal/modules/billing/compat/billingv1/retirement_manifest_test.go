package billingv1

import (
	"os"
	"path/filepath"
	"testing"
)

func readRetirementPair(t *testing.T, name string) []byte {
	t.Helper()
	base := filepath.Join(repositoryRoot(t), "core", "internal", "modules",
		"billing", "migrations", "platform")
	fresh, err := os.ReadFile(filepath.Join(base, "fresh", name+".fresh.sql"))
	if err != nil || len(fresh) == 0 {
		t.Fatalf("Billing V1 retirement migration %s did not load: %v", name, err)
	}
	return fresh
}
