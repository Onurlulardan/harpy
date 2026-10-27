package tenancy

import (
	"strings"
	"testing"
)

func TestTenantActivationV2MigrationLoadsFreshAndChainsCutoverV2(t *testing.T) {
	source := freshMigrationDocument(t, "20260728103010")
	for _, marker := range []string{
		"CREATE FUNCTION platform.tenancy_activation_initialize_v2(",
		"platform.tenancy_activation_initialize_v1(",
		"platform.cutover_initialize_tenant_v2(",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v2(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("tenant activation v2 migration lacks %q", marker)
		}
	}
	if strings.Contains(source, "CREATE OR REPLACE") ||
		strings.Contains(source, "-- +goose Down") {
		t.Fatal("tenant activation v2 migration is not immutable and forward-only")
	}
}
