package postgres

import _ "embed"

//go:embed adapters/postgres/sql/admit_analytics_public.sql
var admitAnalyticsPublicSQL string

//go:embed adapters/postgres/sql/current_bound_schema.sql
var currentBoundSchemaSQL string

//go:embed adapters/postgres/sql/preflight_catalog.sql
var preflightCatalogSQL string

//go:embed adapters/postgres/sql/preflight_roles.sql
var preflightRolesSQL string

//go:embed adapters/postgres/sql/session_hygiene.sql
var sessionHygieneSQL string

//go:embed adapters/postgres/sql/probe_schema_contract.sql
var probeSchemaContractSQL string

//go:embed adapters/postgres/sql/preflight_foundation.sql
var preflightFoundationSQL string

//go:embed adapters/postgres/sql/probe_runtime_role.sql
var probeRuntimeRoleSQL string

//go:embed adapters/postgres/sql/probe_owner_role.sql
var probeOwnerRoleSQL string

//go:embed adapters/postgres/sql/probe_backup_role.sql
var probeBackupRoleSQL string

//go:embed adapters/postgres/sql/resolve_tenant.sql
var resolveTenantSQL string

//go:embed adapters/postgres/sql/resolve_tenant_snapshot.sql
var resolveTenantSnapshotSQL string

//go:embed adapters/postgres/sql/resolve_tenant_slug.sql
var resolveTenantSlugSQL string

//go:embed adapters/postgres/sql/set_tenant_search_path.sql
var setTenantSearchPathSQL string

//go:embed adapters/postgres/sql/verify_tenant_search_path.sql
var verifyTenantSearchPathSQL string

//go:embed adapters/postgres/sql/verify_tenant_identity.sql
var verifyTenantIdentitySQL string

//go:embed adapters/postgres/sql/lock_cutover_group.sql
var lockCutoverGroupSQL string

//go:embed adapters/postgres/sql/list_cutover_group_members.sql
var listCutoverGroupMembersSQL string

//go:embed adapters/postgres/sql/lock_domain_cutover.sql
var lockDomainCutoverSQL string

//go:embed adapters/postgres/sql/verify_domain_cutover.sql
var verifyDomainCutoverSQL string

//go:embed adapters/postgres/sql/resolve_snapshot_domains.sql
var resolveSnapshotDomainsSQL string

//go:embed adapters/postgres/sql/resolve_analytics_public_route.sql
var resolveAnalyticsPublicRouteSQL string

//go:embed adapters/postgres/sql/resolve_analytics_legacy_public.sql
var resolveAnalyticsLegacyPublicSQL string

//go:embed adapters/postgres/sql/record_analytics_legacy_public_hit.sql
var recordAnalyticsLegacyPublicHitSQL string

//go:embed adapters/postgres/sql/resolve_environment_id.sql
var resolveEnvironmentIDSQL string

//go:embed adapters/postgres/sql/resolve_global_billing_generations.sql
var resolveGlobalBillingGenerationsSQL string

//go:embed adapters/postgres/sql/resolve_tenant_billing_generations.sql
var resolveTenantBillingGenerationsSQL string

//go:embed adapters/postgres/sql/resolve_tenant_analytics_generations.sql
var resolveTenantAnalyticsGenerationsSQL string

//go:embed adapters/postgres/sql/resolve_tenant_workflow_generations.sql
var resolveTenantWorkflowGenerationsSQL string

//go:embed adapters/postgres/sql/resolve_tenant_work_generations.sql
var resolveTenantWorkGenerationsSQL string

//go:embed adapters/postgres/sql/resolve_tenant_iam_generations.sql
var resolveTenantIAMGenerationsSQL string

//go:embed adapters/postgres/sql/set_global_search_path.sql
var setGlobalSearchPathSQL string

//go:embed adapters/postgres/sql/verify_global_search_path.sql
var verifyGlobalSearchPathSQL string

//go:embed adapters/postgres/sql/verify_global_environment_id.sql
var verifyGlobalEnvironmentIDSQL string

//go:embed adapters/postgres/sql/lock_global_billing_group.sql
var lockGlobalBillingGroupSQL string

//go:embed adapters/postgres/sql/verify_global_billing_group.sql
var verifyGlobalBillingGroupSQL string

//go:embed adapters/postgres/sql/lock_global_billing_members.sql
var lockGlobalBillingMembersSQL string

//go:embed adapters/postgres/sql/verify_global_billing_members.sql
var verifyGlobalBillingMembersSQL string

//go:embed adapters/postgres/sql/lock_global_billing_domain.sql
var lockGlobalBillingDomainSQL string

//go:embed adapters/postgres/sql/verify_global_billing_domain.sql
var verifyGlobalBillingDomainSQL string

//go:embed adapters/postgres/sql/iam_throttle_blocked.sql
var iamThrottleBlockedSQL string

//go:embed adapters/postgres/sql/iam_throttle_record_failure.sql
var iamThrottleRecordFailureSQL string

//go:embed adapters/postgres/sql/release_analytics_public.sql
var releaseAnalyticsPublicSQL string
