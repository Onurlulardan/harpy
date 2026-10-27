package tenancy

import (
	"path"
	"strings"
	"testing"
)

func TestTenantActivationV4MigrationIsPairedAndExpandCompatible(t *testing.T) {
	const version = "20260729100004"
	pair := map[string]string{}
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasPrefix(name, version+"_") {
			continue
		}
		origin := "fresh"
		if strings.Contains(name, ".v1_adopted.sql") {
			origin = "v1_adopted"
		}
		pair[origin] = string(document.Content())
	}
	if len(pair) != 2 || pair["fresh"] != pair["v1_adopted"] {
		t.Fatal("tenant activation v4 migration is not byte-identical and paired")
	}
	source := pair["fresh"]
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
