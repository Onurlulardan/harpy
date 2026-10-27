package tenancy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCustomerLifecycleMigrationsLoadFreshAndStayNarrow(t *testing.T) {
	names := []string{
		"20260802232300_create_customer_lifecycle_storage_v1",
		"20260802232301_create_customer_lifecycle_reads_v1",
		"20260802232302_create_customer_lifecycle_mutate_v1",
		"20260802232303_create_customer_lifecycle_operations_v1",
		"20260802232304_create_customer_lifecycle_worker_v1",
		"20260802232305_grant_customer_lifecycle_routines_v1",
	}
	for _, name := range names {
		fresh, err := os.ReadFile(filepath.Join("migrations", "platform", "fresh",
			name+".fresh.sql"))
		if err != nil {
			t.Fatal(err)
		}
		if len(fresh) == 0 {
			t.Fatalf("fresh customer lifecycle migration is empty: %s", name)
		}
	}
	storage := lifecycleMigrationText(t, names[0])
	for _, marker := range []string{
		"lifecycle_revision", "tenancy_customer_lifecycle_receipts",
		"tenancy_customer_lifecycle_operations",
		"tenancy_guard_customer_lifecycle_operation_v1",
		"platform.tenancy_customer_lifecycle_operations\nFROM PUBLIC,harpy_runtime",
	} {
		if !strings.Contains(storage, marker) {
			t.Fatalf("customer lifecycle storage missing %q", marker)
		}
	}
	worker := lifecycleMigrationText(t, names[4])
	for _, marker := range []string{
		"FOR UPDATE SKIP LOCKED", "executor_receipt_sha256",
		"PURGE_EXECUTION_FAILED", "UNKNOWN", "proof_verified_at",
	} {
		if !strings.Contains(worker, marker) {
			t.Fatalf("customer lifecycle worker missing %q", marker)
		}
	}
}

func lifecycleMigrationText(t *testing.T, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("migrations", "platform", "fresh",
		name+".fresh.sql"))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
