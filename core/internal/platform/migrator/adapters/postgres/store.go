package postgres

import (
	goosedb "github.com/pressly/goose/v3/database"
)

const historyTableName = "platform.platform_migration_history"

type Store struct {
	metadataByVersion map[int64]migrationMetadata
}

var _ goosedb.StoreExtender = (*Store)(nil)

func (store *Store) Tablename() string {
	return historyTableName
}
