package billing

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func pairedBillingAuthorityMigration(t *testing.T, name string) string {
	t.Helper()
	fresh := readBillingSnapshotMigration(t, "fresh", name+".fresh.sql")
	if len(fresh) == 0 {
		t.Fatalf("%s fresh migration is empty", name)
	}
	if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
		t.Fatalf("%s migration is invalid: %v", name, err)
	}
	return string(fresh)
}
