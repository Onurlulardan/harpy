package notifications

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotificationCutoverMigrationRangeAndSoleGrant(t *testing.T) {
	directory := filepath.Join("migrations", "tenant", "fresh")
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	var adopterGrants int
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "202607182303") {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".fresh.sql") {
			t.Fatalf("cutover migration %s is not fresh sql", entry.Name())
		}
		names = append(names, entry.Name())
		document, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil || len(document) == 0 {
			t.Fatalf("fresh migration %s did not load: %v", entry.Name(), readErr)
		}
		adopterGrants += strings.Count(string(document), "TO harpy_adopter")
		if strings.Contains(string(document), "GRANT INSERT") ||
			strings.Contains(string(document), "GRANT UPDATE") {
			t.Fatalf("%s restores direct adopter DML", entry.Name())
		}
	}
	if len(names) != 20 {
		t.Fatalf("fresh cutover migration count=%d", len(names))
	}
	for index, name := range names {
		want := fmt.Sprintf("202607182303%02d_", index)
		if !strings.HasPrefix(name, want) {
			t.Fatalf("migration[%d]=%s want prefix %s", index, name, want)
		}
	}
	if adopterGrants != 1 {
		t.Fatalf("adopter grants=%d want sole composite grant", adopterGrants)
	}
	hardening, err := os.ReadFile(filepath.Join(directory, names[17]))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"REVOKE INSERT", "FROM harpy_adopter",
		"notifications_migration_intents FROM harpy_runtime",
		"notifications_v1_cutover_receipt_immutable",
	} {
		if !strings.Contains(string(hardening), required) {
			t.Fatalf("hardening missing %q", required)
		}
	}
	grant, err := os.ReadFile(filepath.Join(directory, names[19]))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(grant),
		"notifications_v1_adoption_apply_and_record_v1") ||
		strings.Count(string(grant), "GRANT EXECUTE") != 1 {
		t.Fatal("final migration is not the exact composite grant")
	}
}
