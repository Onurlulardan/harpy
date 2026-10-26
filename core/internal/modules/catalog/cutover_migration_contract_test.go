package catalog

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogCutoverMigrationsArePairedBoundedAndProtected(t *testing.T) {
	fresh, err := filepath.Glob(filepath.Join(
		"migrations", "tenant", "fresh", "202607182305*.fresh.sql",
	))
	if err != nil || len(fresh) != 15 {
		t.Fatalf("Catalog cutover fresh migrations=%d err=%v", len(fresh), err)
	}
	var combined strings.Builder
	for _, freshPath := range fresh {
		body, readErr := os.ReadFile(freshPath)
		if readErr != nil {
			t.Fatal(readErr)
		}
		adoptedName := strings.Replace(
			filepath.Base(freshPath), ".fresh.sql", ".v1_adopted.sql", 1,
		)
		adopted, readErr := os.ReadFile(filepath.Join(
			"migrations", "tenant", "v1_adopted", adoptedName,
		))
		if readErr != nil || !bytes.Equal(body, adopted) {
			t.Fatalf("Catalog cutover pair differs for %s: %v", freshPath, readErr)
		}
		if bytes.Count(body, []byte{'\n'}) > 200 {
			t.Fatalf("Catalog cutover migration exceeds 200 lines: %s", freshPath)
		}
		combined.Write(body)
	}
	lower := strings.ToLower(combined.String())
	for _, required := range []string{
		"catalog.v1_adoption.v1", "platform.cutover_prepare_batch_v1",
		"platform.cutover_record_batch_v1", "session_user",
		"current_user", "tenancy_schema_identity", "origin='v1_adopted'",
		"catalog_adoption_blocker_sha256_v1",
		"catalog_v1_adoption_consumer_sha256_v1",
		"catalog_v1_adoption_receipts", "p_commit_lsn is distinct from p_final_lsn",
	} {
		if !strings.Contains(lower, required) {
			t.Errorf("Catalog cutover contract lacks %q", required)
		}
	}
	for _, forbidden := range []string{
		"insert into modules", "update modules", "delete from modules",
		"execute format", "grant execute on function catalog_v1_adoption_apply_v1",
	} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("Catalog cutover contract contains %q", forbidden)
		}
	}
}

func TestCatalogCutoverGrantExposesOnlyComposite(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "fresh",
		"20260718230513_grant_catalog_cutover_v1.fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(body))
	if strings.Count(text, "grant execute") != 1 ||
		!strings.Contains(text, "catalog_v1_adoption_apply_and_record_v1") {
		t.Fatalf("Catalog adopter grant is not composite-only: %s", text)
	}
}

func TestCatalogCutoverForwardRevokeClosesInternalAdopterSurface(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "fresh",
		"20260718230514_revoke_catalog_adopter_legacy_access_v1.fresh.sql",
	))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(body))
	for _, required := range []string{
		"direct relation privileges are never assigned",
		"catalog_v1_adoption_apply_v1(", "catalog_adoption_inventory_v1(uuid)",
		"catalog_v1_adoption_consumer_sha256_v1(uuid)",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("Catalog adopter forward revoke lacks %q", required)
		}
	}
	if strings.Contains(text, "catalog_v1_adoption_apply_and_record_v1") ||
		strings.Contains(text, "grant execute") {
		t.Fatal("Catalog adopter forward revoke changes the composite-only grant")
	}
}
