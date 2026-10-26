package postgres

import (
	"context"
	"slices"
)

func (transaction *RuntimeTransaction) BindGlobalSearchPath(ctx context.Context) error {
	if !transaction.usable() {
		return ErrTransactionClosed
	}
	var applied string
	if err := transaction.transaction.QueryRow(
		ctx, setGlobalSearchPathSQL,
	).Scan(&applied); err != nil || applied != "pg_catalog" {
		return ErrGlobalIdentityMismatch
	}
	var current []string
	if err := transaction.transaction.QueryRow(
		ctx, verifyGlobalSearchPathSQL,
	).Scan(&current); err != nil || !slices.Equal(current, []string{"pg_catalog"}) {
		return ErrGlobalIdentityMismatch
	}
	return nil
}
