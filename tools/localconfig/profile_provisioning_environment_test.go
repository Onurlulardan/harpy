package main

import "strings"
import "testing"

func TestProvisioningCredentialStaysOffAPIAndWorker(t *testing.T) {
	set, _ := generatedSetForTest(t)
	privileged := []string{
		"HARPY_FOUNDATION_ADMIN_DSN_FILE",
		"HARPY_GOOSE_PASSWORD_FILE",
		"HARPY_API_PASSWORD_FILE",
		"HARPY_WORKER_PASSWORD_FILE",
		"postgres-bootstrap-dsn.txt",
		"postgres-goose-password.txt",
	}
	for _, name := range []string{"api.env", "worker.env"} {
		text := string(set.files[name])
		for _, key := range privileged {
			if strings.Contains(text, key) {
				t.Fatalf("%s contains provisioning credential %s", name, key)
			}
		}
	}
	provisioning := string(set.files["tenant-provisioning.env"])
	for _, key := range []string{
		"HARPY_FOUNDATION_ADMIN_DSN_FILE",
		"HARPY_GOOSE_PASSWORD_FILE",
		"postgres-bootstrap-dsn.txt",
		"postgres-goose-password.txt",
	} {
		if !strings.Contains(provisioning, key) {
			t.Fatalf("tenant-provisioning.env missing %s", key)
		}
	}
	if strings.Contains(provisioning, "HARPY_POSTGRES_DSN_FILE") {
		t.Fatal("tenant provisioning uses the runtime application DSN")
	}
}
