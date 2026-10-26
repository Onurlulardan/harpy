package documents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDocumentsCutoverMigrationRangeAndAuthority(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		directory := filepath.Join("migrations", "tenant", origin)
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		var adopterGrants int
		for _, entry := range entries {
			if !strings.HasPrefix(entry.Name(), "202607182307") {
				continue
			}
			names = append(names, entry.Name())
			document, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
			if readErr != nil {
				t.Fatal(readErr)
			}
			body := string(document)
			adopterGrants += strings.Count(body, "TO harpy_adopter")
			for _, forbidden := range []string{
				"INSERT INTO platform.jobs", "INSERT INTO platform.cutover_job_bindings",
			} {
				if strings.Contains(body, forbidden) {
					t.Fatalf("%s crosses platform owner with %q", entry.Name(), forbidden)
				}
			}
		}
		if len(names) != 20 {
			t.Fatalf("%s cutover migration count=%d", origin, len(names))
		}
		for index, name := range names {
			want := fmt.Sprintf("202607182307%02d_", index+10)
			if !strings.HasPrefix(name, want) {
				t.Fatalf("%s migration[%d]=%s want %s", origin, index, name, want)
			}
		}
		if adopterGrants != 1 {
			t.Fatalf("%s adopter grants=%d want sole composite grant", origin, adopterGrants)
		}
		assertDocumentsHardeningMigration(t, directory, names[17])
		assertDocumentsFinalGrant(t, directory, names[19])
	}
}

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
