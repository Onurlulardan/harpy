package postgres

import (
	"context"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *TenantStore) Delete(
	_ context.Context,
	_ goosedb.DBTxConn,
	version int64,
) error {
	return &DownForbiddenError{Version: version}
}
