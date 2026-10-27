package tenancy

import (
	"strings"
	"testing"
)

func TestTenantActivationV4MigrationLoadsFreshAndStaysExpandCompatible(t *testing.T) {
	source := freshMigrationDocument(t, "20260729100004")
	for _, marker := range []string{
		"CREATE FUNCTION platform.tenancy_activation_initialize_v4(",
		"p_operation_kind text",
		"v_operation.operation_kind IS DISTINCT FROM p_operation_kind",
		"v_application = 'harpy-v2:core-operator-initialize'",
		"v_application = 'harpy-v2:core-worker'",
		"platform.cutover_initialize_tenant_v3(",
		"platform.workflow_register_tenant_authority_v1(",
		"platform.analytics_register_tenant_authority_v1(",
		"GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v4(",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("tenant activation v4 migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"v_application IN (", "CREATE OR REPLACE", "-- +goose Down",
		"REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v1(",
		"REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v2(",
		"REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v3(",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("tenant activation v4 migration contains %q", forbidden)
		}
	}
}
