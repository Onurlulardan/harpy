package tenancy

import (
	"strings"
	"testing"
)

func TestTenantActivationV3MigrationLoadsFreshAndRegistersAuthorities(t *testing.T) {
	source := freshMigrationDocument(t, "20260729100002")
	for _, marker := range []string{
		"CREATE FUNCTION platform.tenancy_activation_initialize_v3(",
		"platform.tenancy_activation_initialize_v2(",
		"platform.analytics_register_tenant_authority_v1(",
		"p_operation_id::text, '7', 15, 1",
		"tenant.created_at INTO STRICT v_created_at",
		"'harpy-v2:core-operator-initialize'",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v3(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("tenant activation v3 migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"CREATE OR REPLACE", "pg_catalog.uuidv7()", "-- +goose Down",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("tenant activation v3 migration contains %q", forbidden)
		}
	}
}
