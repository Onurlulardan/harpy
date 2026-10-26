package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func failJobAttempt(
	ctx context.Context,
	transaction pgx.Tx,
	claim JobClaim,
	failure JobFailure,
	decision jobFailureDecision,
	receiptID any,
	effectStartedAt any,
) (time.Time, error) {
	var occurredAt time.Time
	err := transaction.QueryRow(
		ctx, failJobAttemptSQL,
		claim.attemptID, claim.jobID, claim.leaseOwnerID,
		claim.leaseGeneration, decision.outcome, string(failure.errorClass),
		failure.code, failure.digest, receiptID, effectStartedAt,
	).Scan(&occurredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobLeaseLost
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("fail job attempt: %w", err)
	}
	return occurredAt, nil
}
