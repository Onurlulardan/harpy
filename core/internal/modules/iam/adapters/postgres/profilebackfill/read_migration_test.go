package profilebackfill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticPublicationReadsOnlyCallGuardedRoutines(t *testing.T) {
	for name, document := range map[string]string{
		"source": readSQL, "receipt": readReceiptSQL,
	} {
		for _, table := range []string{
			"iam_users", "iam_organizations", "iam_roles",
			"iam_authorization_operation_receipts",
		} {
			if strings.Contains(document, table) {
				t.Fatalf("%s static SQL reads raw table %s", name, table)
			}
		}
	}
	if !strings.Contains(readSQL, "iam_profile_backfill_source_resources_plan_v1(") ||
		!strings.Contains(readReceiptSQL, "iam_profile_publication_receipt_read_v1(") {
		t.Fatal("static SQL does not call guarded publication routines")
	}
}

func readProfileBackfillMigration(t *testing.T, origin, name string) string {
	t.Helper()
	path := filepath.Join("..", "..", "..", "migrations", "tenant", origin,
		name+"."+origin+".sql")
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(document)
}
