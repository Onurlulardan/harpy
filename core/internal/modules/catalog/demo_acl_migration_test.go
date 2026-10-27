package catalog

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogDemoOperationPlanACLRepairIsExactAndPaired(t *testing.T) {
	const name = "20260802232200_repair_catalog_demo_operation_plan_acl_v1"
	fresh, err := os.ReadFile(filepath.Join(
		"migrations", "tenant", "fresh", name+".fresh.sql",
	))
	if err != nil || len(fresh) == 0 {
		t.Fatalf("Catalog demo operation-plan ACL repair did not load: %v", err)
	}
	if !bytes.Contains(fresh, []byte("REVOKE INSERT ON TABLE catalog_demo_operation_plans")) ||
		!bytes.Contains(fresh, []byte("GRANT INSERT (")) ||
		!bytes.Contains(fresh, []byte("created_at\n) ON TABLE")) {
		t.Fatal("Catalog demo operation-plan ACL repair drifted")
	}
}
