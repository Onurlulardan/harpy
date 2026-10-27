package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (transaction *RuntimeTransaction) VerifyGlobalEnvironmentID(
	ctx context.Context,
	expected uuid.UUID,
) error {
	if !transaction.usable() || expected == uuid.Nil {
		return ErrGlobalIdentityMismatch
	}
	var matches bool
	if err := transaction.transaction.QueryRow(
		ctx, verifyGlobalEnvironmentIDSQL, expected,
	).Scan(&matches); err != nil || !matches {
		return ErrGlobalIdentityMismatch
	}
	return nil
}
