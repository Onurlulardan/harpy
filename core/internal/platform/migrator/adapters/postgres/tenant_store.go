package postgres

import goosedb "github.com/pressly/goose/v3/database"

const tenantCheckpointTableName = "platform.tenant_migration_checkpoints"

type TenantStore struct {
	target            TenantMigrationTarget
	session           TenantSessionProof
	metadataByVersion map[int64]tenantMigrationMetadata
	finalVersion      int64
}

var _ goosedb.StoreExtender = (*TenantStore)(nil)

func (store *TenantStore) Tablename() string {
	return tenantCheckpointTableName
}
