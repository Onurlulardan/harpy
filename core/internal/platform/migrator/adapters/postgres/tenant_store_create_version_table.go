package postgres

import (
	"context"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *TenantStore) CreateVersionTable(
	_ context.Context,
	_ goosedb.DBTxConn,
) error {
	return fmt.Errorf("tenant checkpoint infrastructure must be platform-migrated first")
}
