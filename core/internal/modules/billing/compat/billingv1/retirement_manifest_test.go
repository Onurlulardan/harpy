package billingv1

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRetirementManifestMatchesProvisionedInventory(t *testing.T) {
	t.Parallel()
	first := readRetirementPair(t,
		"20260717010240_provision_billing_v1_http_retirements_v1")
	second := readRetirementPair(t,
		"20260717010241_assert_billing_v1_http_retirements_v1")
	provisioning := string(first) + string(second)
	if strings.Count(provisioning, "'rq.alias.") != 26 {
		t.Fatalf("provisioned key count = %d", strings.Count(provisioning, "'rq.alias."))
	}
	manifest, err := RetirementManifest()
	if err != nil {
		t.Fatal(err)
	}
	identity, keys, err := manifest.DatabaseValues()
	if err != nil || len(keys) != 26 {
		t.Fatalf("manifest keys=%d error=%v", len(keys), err)
	}
	owner, domain, err := identity.DatabaseValues()
	if err != nil || owner != "billing_v1_http" || domain != "billing" {
		t.Fatalf("manifest identity=%s/%s error=%v", owner, domain, err)
	}
}

func TestRetirementMigrationsUseSupportedTopLevelCalls(t *testing.T) {
	t.Parallel()
	first := string(readRetirementPair(t,
		"20260717010240_provision_billing_v1_http_retirements_v1"))
	second := string(readRetirementPair(t,
		"20260717010241_assert_billing_v1_http_retirements_v1"))
	for index, migration := range []string{first, second} {
		if strings.Contains(migration, "DO $") || strings.Count(migration,
			"SELECT platform.contract_retirement_provision_v1(") != 13 {
			t.Fatalf("retirement migration %d must contain 13 top-level calls", index)
		}
	}
}

func readRetirementPair(t *testing.T, name string) []byte {
	t.Helper()
	base := filepath.Join(repositoryRoot(t), "core", "internal", "modules",
		"billing", "migrations", "platform")
	fresh, err := os.ReadFile(filepath.Join(base, "fresh", name+".fresh.sql"))
	if err != nil || len(fresh) == 0 {
		t.Fatalf("Billing V1 retirement migration %s did not load: %v", name, err)
	}
	return fresh
}
