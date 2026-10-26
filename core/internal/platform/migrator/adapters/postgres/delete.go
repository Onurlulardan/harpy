package postgres

import (
	"context"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *Store) Delete(
	_ context.Context,
	_ goosedb.DBTxConn,
	version int64,
) error {
	return &DownForbiddenError{Version: version}
}
