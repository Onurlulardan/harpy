package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (transaction *RuntimeTransaction) ResolveGlobalEnvironmentID(
	ctx context.Context,
) (uuid.UUID, error) {
	if !transaction.usable() {
		return uuid.Nil, ErrTransactionClosed
	}
	var environmentID uuid.UUID
	if err := transaction.transaction.QueryRow(
		ctx, resolveEnvironmentIDSQL,
	).Scan(&environmentID); err != nil || environmentID == uuid.Nil {
		return uuid.Nil, ErrGlobalIdentityMismatch
	}
	return environmentID, nil
}
