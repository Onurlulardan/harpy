package postgres

import (
	"context"
	"database/sql"
	"fmt"

	goosedb "github.com/pressly/goose/v3/database"
)

func (store *Store) CreateVersionTable(
	ctx context.Context,
	db goosedb.DBTxConn,
) error {
	transaction, ok := db.(*sql.Tx)
	if !ok || transaction == nil {
		return ErrCreateTransactionRequired
	}
	statements := []struct {
		name string
		sql  string
	}{
		{"platform schema", createSchemaSQL},
		{"migration history table", createHistoryTableSQL},
		{"mutation rejection function", createRejectMutationFunctionSQL},
		{"mutation rejection trigger", createRejectMutationTriggerSQL},
	}
	for _, statement := range statements {
		if _, err := transaction.ExecContext(ctx, statement.sql); err != nil {
			return fmt.Errorf("create %s: %w", statement.name, err)
		}
	}
	return nil
}
