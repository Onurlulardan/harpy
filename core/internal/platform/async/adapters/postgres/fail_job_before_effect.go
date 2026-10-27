package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *repository) FailJobBeforeEffect(
	ctx context.Context,
	claim JobClaim,
	failure JobFailure,
) (JobFailureResult, error) {
	if ctx == nil || repository == nil ||
		!repository.authority.matchesClaim(claim) {
		return JobFailureResult{}, ErrAuthorityMismatch
	}
	decision, err := decideJobFailure(claim, failure)
	if err != nil {
		return JobFailureResult{}, err
	}
	locked, err := repository.lockClaim(ctx, claim)
	if err != nil {
		return JobFailureResult{}, err
	}
	if locked.effectStarted.Valid {
		return JobFailureResult{}, ErrExternalEffectPermit
	}
	var occurredAt time.Time
	err = repository.database.QueryRow(
		ctx, failJobAttemptSQL,
		claim.attemptID, claim.jobID, claim.leaseOwnerID,
		claim.leaseGeneration, decision.outcome, string(failure.errorClass),
		failure.code, failure.digest, nil, nil,
	).Scan(&occurredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return JobFailureResult{}, ErrJobLeaseLost
	}
	if err != nil {
		return JobFailureResult{}, fmt.Errorf("fail bound job attempt: %w", err)
	}
	return repository.persistFailedClaim(ctx, claim, failure, decision, occurredAt)
}

func (repository *repository) persistFailedClaim(
	ctx context.Context,
	claim JobClaim,
	failure JobFailure,
	decision jobFailureDecision,
	occurredAt time.Time,
) (JobFailureResult, error) {
	arguments := append(
		authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
		string(decision.state), string(failure.errorClass), failure.code,
		failure.digest, failure.retryDelay.Microseconds(), occurredAt, nil,
	)
	var persistedAt time.Time
	var retryAt pgtype.Timestamptz
	err := repository.database.QueryRow(ctx, failJobSQL, arguments...).Scan(
		&persistedAt, &retryAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return JobFailureResult{}, ErrJobLeaseLost
	}
	if err != nil {
		return JobFailureResult{}, fmt.Errorf("persist bound failed job: %w", err)
	}
	result := JobFailureResult{state: decision.state, occurredAt: persistedAt}
	if retryAt.Valid {
		result.runAt = retryAt.Time
	}
	return result, nil
}
