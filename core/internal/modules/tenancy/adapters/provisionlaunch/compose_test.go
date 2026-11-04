package provisionlaunch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComposeArgsUseFixedTenantProvisioningService(t *testing.T) {
	args := strings.Join(composeArgs("00112233-4455-6677-8899-aabbccddeeff"), " ")
	if !strings.Contains(args, "run --rm --no-deps") ||
		!strings.Contains(args, "-f ops/local/compose.runtime.yml") ||
		!strings.HasSuffix(args, "tenant-provisioning") ||
		strings.Contains(args, "docker.sock") {
		t.Fatal(args)
	}
	if strings.Count(args, "HARPY_PROVISION_TENANT_ID=00112233-4455-6677-8899-aabbccddeeff") != 1 {
		t.Fatal(args)
	}
}

func TestRuntimeComposeKeepsProvisioningOneShot(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(moduleRoot(t), "ops/local/compose.runtime.yml"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "restart: \"no\"") ||
		!strings.Contains(text, "env_file: ./state/secrets/tenant-provisioning.env") ||
		!strings.Contains(text, "entrypoint: [\"/app/core-initialize\"]") {
		t.Fatal("tenant-provisioning is not a one-shot initialize service")
	}
	// PRESERVE_NON_MIGRATION: absence guard for removed executor service names.
	for _, forbidden := range []string{
		"HARPY_GOOSE_PASSWORD", "HARPY_FOUNDATION_ADMIN",
		"tenant-migration-executor", "schema-executor", "core-migrator",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("compose.runtime.yml contains %s", forbidden)
		}
	}
	if strings.Count(text, "env_file: ./state/secrets/api.env") != 1 ||
		strings.Count(text, "env_file: ./state/secrets/worker.env") != 1 {
		t.Fatal(text)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
