package analytics

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyticsCutoverMigrationsArePairedBoundedAndProtected(t *testing.T) {
	platform := analyticsCutoverFiles(t, "platform", "202607182306*.fresh.sql", 1)
	tenant := analyticsCutoverFiles(t, "tenant", "202607182306*.fresh.sql", 19)
	platformText := analyticsCutoverPairs(t, "platform", platform)
	tenantText := analyticsCutoverPairs(t, "tenant", tenant)

	for _, required := range []string{
		"analytics.v1_adoption.v1", "platform.cutover_prepare_batch_v1",
		"platform.cutover_record_batch_v1", "session_user", "current_user",
		"origin='v1_adopted'", "analytics_v1_adoption_receipts",
		"analytics_v1_adoption_apply_and_record_v1",
		"platform.analytics_v1_adoption_public_bridge_v1",
	} {
		if !strings.Contains(tenantText+platformText, required) {
			t.Errorf("Analytics cutover contract lacks %q", required)
		}
	}
	for _, forbidden := range []string{
		"platform.analytics_publications", "platform.analytics_legacy_public_routes",
		"grant execute on function analytics_v1_adoption_apply_v1",
	} {
		if strings.Contains(tenantText, forbidden) {
			t.Errorf("Analytics tenant cutover contains %q", forbidden)
		}
	}
	if strings.Count(tenantText, "grant execute") != 1 ||
		!strings.Contains(tenantText,
			"grant execute on function analytics_v1_adoption_apply_and_record_v1") {
		t.Fatal("Analytics adopter grant is not composite-only")
	}
	for _, required := range []string{
		"revoke execute on function platform.analytics_register_tenant_authority_v1",
		"revoke execute on function platform.analytics_v1_adoption_public_bridge_v1",
	} {
		if !strings.Contains(platformText, required) {
			t.Errorf("Analytics platform cutover lacks %q", required)
		}
	}
}

func analyticsCutoverFiles(t *testing.T, scope, pattern string, want int) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("migrations", scope, "fresh", pattern))
	if err != nil || len(paths) != want {
		t.Fatalf("Analytics %s cutover migrations=%d want=%d err=%v",
			scope, len(paths), want, err)
	}
	return paths
}

func analyticsCutoverPairs(t *testing.T, scope string, fresh []string) string {
	t.Helper()
	var combined strings.Builder
	for _, freshPath := range fresh {
		body, err := os.ReadFile(freshPath)
		if err != nil {
			t.Fatal(err)
		}
		adoptedName := strings.Replace(
			filepath.Base(freshPath), ".fresh.sql", ".v1_adopted.sql", 1,
		)
		adopted, readErr := os.ReadFile(filepath.Join(
			"migrations", scope, "v1_adopted", adoptedName,
		))
		if readErr != nil || !bytes.Equal(body, adopted) {
			t.Fatalf("Analytics cutover pair differs for %s: %v", freshPath, readErr)
		}
		if bytes.Count(body, []byte{'\n'}) > 200 {
			t.Fatalf("Analytics cutover migration exceeds 200 lines: %s", freshPath)
		}
		combined.Write(body)
	}
	return strings.ToLower(combined.String())
}
