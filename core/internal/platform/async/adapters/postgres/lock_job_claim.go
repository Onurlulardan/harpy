package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type lockedJobClaim struct {
	cancelRequested pgtype.Timestamptz
	effectStarted   pgtype.Timestamptz
	leaseUntil      time.Time
}

func lockJobClaim(
	ctx context.Context,
	transaction pgx.Tx,
	claim JobClaim,
) (lockedJobClaim, error) {
	var locked lockedJobClaim
	arguments := append(
		authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
	)
	err := transaction.QueryRow(ctx, lockJobClaimSQL, arguments...).Scan(
		&locked.cancelRequested,
		&locked.effectStarted,
		&locked.leaseUntil,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return lockedJobClaim{}, ErrJobLeaseLost
	}
	if err != nil {
		return lockedJobClaim{}, fmt.Errorf("lock active job claim: %w", err)
	}
	return locked, nil
}
