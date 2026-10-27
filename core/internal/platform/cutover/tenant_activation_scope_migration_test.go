package cutover

import (
	"strings"
	"testing"
)

func TestTenantActivationScopeMigrationIsPairedAndFailClosed(t *testing.T) {
	const version = "20260729100003"
	source := tenantSingletonAuthorityMigrationPair(t, version)["fresh"]
	for _, marker := range []string{
		"CREATE FUNCTION platform.cutover_initialize_tenant_v3(",
		"p_operation_kind text",
		"p_operation_kind = 'environment_initialize'",
		"v_application = 'harpy-v2:core-operator-initialize'",
		"p_operation_kind = 'saas_customer_tenant_provision'",
		"v_application = 'harpy-v2:core-worker'",
		"platform.cutover_initialize_tenant_v2(",
		"tenant cutover initialization v3 authority conflict",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("tenant activation scope migration lacks %q", marker)
		}
	}
	for _, forbidden := range []string{
		"v_application IN (",
		"GRANT EXECUTE ON FUNCTION platform.cutover_initialize_tenant_v3(",
		"CREATE OR REPLACE",
		"-- +goose Down",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("tenant activation scope migration contains %q", forbidden)
		}
	}
}
