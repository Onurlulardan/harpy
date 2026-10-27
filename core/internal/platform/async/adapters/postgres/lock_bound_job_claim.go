package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (repository *repository) lockClaim(
	ctx context.Context,
	claim JobClaim,
) (lockedJobClaim, error) {
	var locked lockedJobClaim
	arguments := append(
		authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
	)
	err := repository.database.QueryRow(ctx, lockJobClaimSQL, arguments...).Scan(
		&locked.cancelRequested,
		&locked.effectStarted,
		&locked.leaseUntil,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return lockedJobClaim{}, ErrJobLeaseLost
	}
	if err != nil {
		return lockedJobClaim{}, fmt.Errorf("lock bound job claim: %w", err)
	}
	return locked, nil
}
