package postgres

import _ "embed"

//go:embed sql/resolve_tenant_schema.sql
var schemaExecutorResolveTenantSQL string

//go:embed sql/acquire_tenant_lock.sql
var schemaExecutorAcquireTenantLockSQL string

//go:embed sql/release_tenant_lock.sql
var schemaExecutorReleaseTenantLockSQL string

//go:embed sql/verify_tenant_sentinel.sql
var schemaExecutorVerifySentinelSQL string

//go:embed sql/set_tenant_search_path.sql
var schemaExecutorSetSearchPathSQL string

//go:embed sql/read_relation_owner.sql
var schemaExecutorReadRelationOwnerSQL string

//go:embed sql/read_column_contract.sql
var schemaExecutorReadColumnSQL string

//go:embed sql/read_index_contract.sql
var schemaExecutorReadIndexSQL string

//go:embed sql/read_constraint_contract.sql
var schemaExecutorReadConstraintSQL string

//go:embed sql/read_named_constraint.sql
var schemaExecutorReadNamedConstraintSQL string

//go:embed sql/resolve_trigram_schema.sql
var schemaExecutorResolveTrigramSchemaSQL string
