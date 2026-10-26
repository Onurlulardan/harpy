package runtimev1

import (
	"encoding/json"
	"os"
	"testing"
)

type phase14ConfigCatalog struct {
	Profiles  []string                `json:"profiles"`
	Variables []phase14ConfigVariable `json:"variables"`
}

type phase14ConfigVariable struct {
	Name          string            `json:"name"`
	Access        string            `json:"access"`
	ProfileAccess map[string]string `json:"profile_access"`
}

func TestPhase14CatalogPublishesExactArtifactAccess(t *testing.T) {
	catalog := readPhase14ConfigCatalog(t)
	variables := make(map[string]phase14ConfigVariable, len(catalog.Variables))
	for _, variable := range catalog.Variables {
		if _, duplicate := variables[variable.Name]; duplicate {
			t.Fatalf("duplicate Phase14 variable %s", variable.Name)
		}
		variables[variable.Name] = variable
	}
	expected := map[string]string{
		"HARPY_INITIAL_ADMIN_PASSWORD_FILE":        "read_once_cleanup",
		"HARPY_RECOVERY_SIGNING_DIRECTORY":         "create_only",
		"HARPY_LOCAL_AUTHORITY_HEAD_JOURNAL_PATH":  "read_write",
		"HARPY_RESTORE_PROGRESS_PATH":              "atomic_write",
		"HARPY_RESTORE_DRAIN_ACK_PATH":             "read_only",
		"HARPY_RECOVERY_KIT_FILE":                  "read_once_cleanup",
		"HARPY_RECOVERY_MANIFEST_SIGNING_KEY_FILE": "read_only",
		"HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE":   "read_only",
	}
	for key, access := range expected {
		if variables[key].Access != access {
			t.Fatalf("%s access=%q, want %q", key, variables[key].Access, access)
		}
	}
	receiptAccess := variables["HARPY_RESTORE_RECEIPT_PATH"].ProfileAccess
	if receiptAccess["database_restore"] != "atomic_write" ||
		receiptAccess["operator_identity_reconcile"] != "read_only" ||
		receiptAccess["operator_license_reconcile"] != "read_only" {
		t.Fatalf("restore receipt writer/reader split drifted: %#v", receiptAccess)
	}
}

func readPhase14ConfigCatalog(t *testing.T) phase14ConfigCatalog {
	t.Helper()
	content, err := os.ReadFile("phase14-config.json")
	if err != nil {
		t.Fatal(err)
	}
	var catalog phase14ConfigCatalog
	if err := json.Unmarshal(content, &catalog); err != nil {
		t.Fatal(err)
	}
	return catalog
}
