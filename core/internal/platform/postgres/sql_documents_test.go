package postgres

import (
	"strings"
	"testing"
)

func TestRuntimeSQLDocumentsAreEmbedded(t *testing.T) {
	documents := []string{
		admitAnalyticsPublicSQL,
		preflightCatalogSQL, preflightRolesSQL, sessionHygieneSQL,
		probeSchemaContractSQL, resolveTenantSQL, resolveTenantSnapshotSQL,
		setTenantSearchPathSQL,
		verifyTenantSearchPathSQL, verifyTenantIdentitySQL,
		lockCutoverGroupSQL, listCutoverGroupMembersSQL, lockDomainCutoverSQL,
		verifyDomainCutoverSQL,
		resolveSnapshotDomainsSQL, resolveAnalyticsPublicRouteSQL,
		resolveAnalyticsLegacyPublicSQL, recordAnalyticsLegacyPublicHitSQL,
		resolveEnvironmentIDSQL, setGlobalSearchPathSQL, verifyGlobalSearchPathSQL,
		verifyGlobalEnvironmentIDSQL, lockGlobalBillingGroupSQL,
		verifyGlobalBillingGroupSQL, lockGlobalBillingMembersSQL,
		verifyGlobalBillingMembersSQL, lockGlobalBillingDomainSQL,
		verifyGlobalBillingDomainSQL,
		resolveTenantAnalyticsGenerationsSQL,
		resolveTenantCatalogGenerationsSQL,
		iamThrottleBlockedSQL, iamThrottleRecordFailureSQL,
		releaseAnalyticsPublicSQL,
	}
	for index, document := range documents {
		if strings.TrimSpace(document) == "" || !strings.HasSuffix(document, "\n") {
			t.Fatalf("runtime SQL document %d is empty or not newline terminated", index)
		}
	}
}

func TestCurrentBoundSchemaUsesProductionTenantSearchPath(t *testing.T) {
	if !strings.Contains(currentBoundSchemaSQL, "current_schemas(false)") ||
		!strings.Contains(currentBoundSchemaSQL, "schemas[1]::text = 'pg_catalog'") ||
		!strings.Contains(currentBoundSchemaSQL, "schemas[2]::text") ||
		strings.Contains(currentBoundSchemaSQL, "current_schema()") {
		t.Fatal("current bound schema must read the verified pg_catalog,tenant path")
	}
}

func TestCommandLocksUseOwnerRoutines(t *testing.T) {
	documents := []struct {
		document string
		routine  string
		direct   string
	}{
		{resolveTenantSQL, "tenancy_runtime_resolve_command_v1", "FROM platform.tenants"},
		{lockCutoverGroupSQL, "cutover_runtime_lock_group_v1", "FROM platform.cutover_groups"},
		{listCutoverGroupMembersSQL, "cutover_runtime_lock_group_members_v1", "FROM platform.cutover_group_members"},
		{lockDomainCutoverSQL, "cutover_runtime_lock_domain_v1", "FROM platform.domain_cutovers"},
		{lockGlobalBillingGroupSQL, "cutover_runtime_lock_group_v1", "FROM platform.cutover_groups"},
		{lockGlobalBillingMembersSQL, "cutover_runtime_lock_group_members_v1", "FROM platform.cutover_group_members"},
		{lockGlobalBillingDomainSQL, "cutover_runtime_lock_domain_v1", "FROM platform.domain_cutovers"},
	}
	for _, value := range documents {
		if !strings.Contains(value.document, value.routine) ||
			strings.Contains(value.document, value.direct) {
			t.Fatalf("runtime lock must cross %s owner routine", value.routine)
		}
	}
}
