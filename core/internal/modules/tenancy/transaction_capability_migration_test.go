package tenancy

import (
	"os"
	"path/filepath"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func freshTenancyMigration(t *testing.T, name string) string {
	t.Helper()
	fresh, err := os.ReadFile(filepath.Join(
		"migrations", "platform", "fresh", name+".fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	if len(fresh) == 0 {
		t.Fatalf("%s fresh migration is empty", name)
	}
	if _, err = postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("%s migration is invalid: %v", name, err)
	}
	return string(fresh)
}
