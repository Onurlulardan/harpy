package documents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertDocumentsHardeningMigration(t *testing.T, directory, name string) {
	t.Helper()
	document, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	body := string(document)
	for _, required := range []string{
		"REVOKE SELECT ON", "FROM harpy_adopter",
		"REVOKE EXECUTE ON FUNCTION documents_guard_folder_cycle_v1()\nFROM harpy_adopter",
		"REVOKE ALL ON FUNCTION documents_v1_receipt_reject_mutation_v1() FROM PUBLIC",
		"REVOKE ALL ON FUNCTION documents_v1_reverse_reconciliation_guard_v1() FROM PUBLIC",
		"documents_v1_cutover_receipt_immutable",
		"documents_v1_reverse_receipt_immutable",
		"documents_v1_reverse_reconciliation_guard",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("%s hardening missing %q", name, required)
		}
	}
}

func assertDocumentsFinalGrant(t *testing.T, directory, name string) {
	t.Helper()
	document, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	body := string(document)
	if !strings.Contains(body, "documents_v1_adoption_apply_and_record_v1") ||
		strings.Count(body, "GRANT EXECUTE") != 1 {
		t.Fatalf("%s is not the exact composite grant", name)
	}
}
