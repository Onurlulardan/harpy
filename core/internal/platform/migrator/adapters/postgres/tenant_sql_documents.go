package postgres

import _ "embed"

//go:embed sql/list_tenant_migration_targets.sql
var listTenantMigrationTargetsSQL string

//go:embed sql/tenant_infrastructure_exists.sql
var tenantInfrastructureExistsSQL string

//go:embed sql/read_tenant_checkpoint.sql
var readTenantCheckpointSQL string

//go:embed sql/lock_tenant_checkpoint.sql
var lockTenantCheckpointSQL string

//go:embed sql/insert_tenant_checkpoint.sql
var insertTenantCheckpointSQL string

//go:embed sql/start_tenant_checkpoint.sql
var startTenantCheckpointSQL string

//go:embed sql/insert_tenant_checkpoint_attempt.sql
var insertTenantCheckpointAttemptSQL string

//go:embed sql/complete_tenant_checkpoint.sql
var completeTenantCheckpointSQL string

//go:embed sql/tenant_checkpoint_table_exists.sql
var tenantCheckpointTableExistsSQL string

//go:embed sql/list_tenant_store_migrations.sql
var listTenantStoreMigrationsSQL string

//go:embed sql/get_tenant_store_migration.sql
var getTenantStoreMigrationSQL string

//go:embed sql/get_latest_tenant_store_version.sql
var getLatestTenantStoreVersionSQL string

//go:embed sql/acquire_tenant_migration_lock.sql
var acquireTenantMigrationLockSQL string

//go:embed sql/release_tenant_migration_lock.sql
var releaseTenantMigrationLockSQL string

//go:embed sql/set_tenant_migration_search_path.sql
var setTenantMigrationSearchPathSQL string

//go:embed sql/read_tenant_migration_session.sql
var readTenantMigrationSessionSQL string

//go:embed sql/verify_tenant_migration_session.sql
var verifyTenantMigrationSessionSQL string

//go:embed sql/insert_tenant_schema_identity.sql
var insertTenantSchemaIdentitySQL string

//go:embed sql/verify_tenant_schema_identity.sql
var verifyTenantSchemaIdentitySQL string
