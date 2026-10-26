package local

import (
	"os"
	"strings"
	"testing"
)

func TestInstallationSecretInventoryIsPreDisplayAndFailClosed(t *testing.T) {
	compose, err := os.ReadFile("compose.inventory.yml")
	if err != nil {
		t.Fatal(err)
	}
	document := string(compose)
	for _, required := range []string{
		"entrypoint: [\"/app/plaintext-inventory\"]",
		"- /run/installation/initial-admin-password",
		"- -allow-absent",
		"- /scan/api",
		"- /scan/worker",
		"- /scan/release",
		"- /scan/dev-authority",
		"- /scan/recovery-signing",
		"- /run/database/postgres-bootstrap-dsn.txt",
		"- pre-display-compose-logs",
	} {
		if !strings.Contains(document, required) {
			t.Fatalf("inventory compose is missing %q", required)
		}
	}
	tasks, err := os.ReadFile("../../.taskfiles/local.yml")
	if err != nil {
		t.Fatal(err)
	}
	initialize := taskBlock(t, string(tasks), "initialize", "services")
	assertOrdered(t, initialize, []string{
		"installation-delivery prepare",
		"run --rm core-operator",
		"task: :local:inventory:verify",
		"installation-delivery deliver",
		"installation-delivery verify",
	})
}

func TestInventoryTaskScansComposeLogsWithoutPersistingThem(t *testing.T) {
	content, err := os.ReadFile("../../.taskfiles/local-inventory.yml")
	if err != nil {
		t.Fatal(err)
	}
	document := string(content)
	if !strings.Contains(document, "logs --no-color |") ||
		!strings.Contains(document, "run --rm -T installation-secret-inventory") ||
		strings.Contains(document, "> ") || strings.Contains(document, "tee ") {
		t.Fatal("inventory task must stream logs directly into the scanner")
	}
}
