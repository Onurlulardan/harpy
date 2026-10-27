package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (transaction *RuntimeTransaction) LockGlobalBillingGroup(
	ctx context.Context,
	environmentID uuid.UUID,
	expectedGeneration int64,
) error {
	return transaction.readGlobalBillingGroup(
		ctx, lockGlobalBillingGroupSQL, environmentID, expectedGeneration,
	)
}

func (transaction *RuntimeTransaction) VerifyGlobalBillingGroup(
	ctx context.Context,
	environmentID uuid.UUID,
	expectedGeneration int64,
) error {
	return transaction.readGlobalBillingGroup(
		ctx, verifyGlobalBillingGroupSQL, environmentID, expectedGeneration,
	)
}

func (transaction *RuntimeTransaction) readGlobalBillingGroup(
	ctx context.Context,
	document string,
	environmentID uuid.UUID,
	expectedGeneration int64,
) error {
	if !transaction.usable() || environmentID == uuid.Nil || expectedGeneration <= 0 {
		return ErrAuthorityMismatch
	}
	var state, writer string
	var generation int64
	err := transaction.transaction.QueryRow(
		ctx, document, environmentID,
	).Scan(&state, &writer, &generation)
	if err != nil || state != "v2" || writer != "v2" || generation != expectedGeneration {
		return ErrAuthorityMismatch
	}
	return nil
}
