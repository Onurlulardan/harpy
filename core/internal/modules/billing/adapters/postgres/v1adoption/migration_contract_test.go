package v1adoption

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestProtectedCutoverMigrationsArePairedBoundedAndSealed(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "platform")
	fresh := billingProtectedPaths(t, filepath.Join(root, "fresh"), ".fresh.sql")
	if len(fresh) != 13 {
		t.Fatalf("Billing protected fresh migration count=%d", len(fresh))
	}
	for index := range fresh {
		left := readBillingProtectedMigration(t, fresh[index])
		if len(left) == 0 {
			t.Fatalf("Billing protected migration is empty: %s", filepath.Base(fresh[index]))
		}
		if lines := bytes.Count(left, []byte("\n")) + 1; lines > 200 {
			t.Fatalf("Billing migration has %d lines: %s", lines, filepath.Base(fresh[index]))
		}
	}
}

func TestProtectedCutoverUsesExactAuthorityAndReceipts(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "platform", "fresh")
	stage := string(readBillingProtectedMigration(t, filepath.Join(root,
		"20260718230405_create_billing_v1_adoption_stage_plan_v1.fresh.sql")))
	for _, token := range []string{
		"platform.applied_change_receipts",
		"(p_batch_ordinal - 1)::bigint * 1000",
		"p_lease_manifest_sha256", "source_xid",
	} {
		if !strings.Contains(stage, token) {
			t.Fatalf("Billing protected stage missing %q", token)
		}
	}
	grant := string(readBillingProtectedMigration(t, filepath.Join(root,
		"20260718230412_grant_billing_v1_adoption_composites_v1.fresh.sql")))
	if strings.Count(grant, "GRANT EXECUTE ON FUNCTION") != 2 ||
		strings.Contains(grant, "_global_apply_v1(") ||
		strings.Contains(grant, "_tenant_apply_v1(") {
		t.Fatalf("Billing adopter grant is wider than two composites: %s", grant)
	}
}

func TestProtectedCutoverRevokesEveryLegacyAdopterRoutine(t *testing.T) {
	root := filepath.Join("..", "..", "..", "migrations", "platform", "fresh")
	document := string(readBillingProtectedMigration(t, filepath.Join(root,
		"20260718230400_revoke_legacy_billing_v1_adoption_access_v1.fresh.sql")))
	for _, routine := range []string{
		"billing_convert_v1_storage_megabytes_v1",
		"billing_register_v1_adoption_v1",
		"billing_apply_v1_adoption_v1",
		"billing_verify_v1_adoption_v1",
		"billing_rehearse_v1_adoption_rollback_v1",
		"billing_assert_v1_adoption_ready_v1",
	} {
		if !strings.Contains(document, routine) {
			t.Fatalf("Billing adopter revocation missing %s", routine)
		}
	}
	if strings.Count(document, "REVOKE EXECUTE ON FUNCTION") != 6 {
		t.Fatalf("Billing adopter legacy revocation differs: %s", document)
	}
}

func billingProtectedPaths(t *testing.T, directory, suffix string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, 0, 13)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "202607182304") &&
			strings.HasSuffix(entry.Name(), suffix) {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	return paths
}

func readBillingProtectedMigration(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read Billing protected migration %s: %v", path, err)
	}
	return content
}
