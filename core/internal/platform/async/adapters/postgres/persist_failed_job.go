package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func persistFailedJob(
	ctx context.Context,
	transaction pgx.Tx,
	claim JobClaim,
	failure JobFailure,
	decision jobFailureDecision,
	occurredAt time.Time,
	effectStartedAt time.Time,
) (JobFailureResult, error) {
	var effectStarted any
	if !effectStartedAt.IsZero() {
		effectStarted = effectStartedAt
	}
	arguments := append(
		authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
		string(decision.state), string(failure.errorClass), failure.code,
		failure.digest, failure.retryDelay.Microseconds(), occurredAt, effectStarted,
	)
	var persistedAt time.Time
	var retryAt pgtype.Timestamptz
	err := transaction.QueryRow(ctx, failJobSQL, arguments...).Scan(
		&persistedAt, &retryAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return JobFailureResult{}, ErrJobLeaseLost
	}
	if err != nil {
		return JobFailureResult{}, fmt.Errorf("persist failed job: %w", err)
	}
	result := JobFailureResult{state: decision.state, occurredAt: persistedAt}
	if retryAt.Valid {
		result.runAt = retryAt.Time
	}
	return result, nil
}
