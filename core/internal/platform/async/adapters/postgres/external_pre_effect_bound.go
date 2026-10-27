package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (store *boundDatabaseJobStore) startExternal(
	ctx context.Context,
	owner *workerIdentity,
	claim JobClaim,
	spec ExternalEffectRequestSpec,
) (ExternalEffectPermit, error) {
	if _, err := buildExternalEffectRequest(claim, spec, claim.startedAt); err != nil {
		return ExternalEffectPermit{}, err
	}
	locked, err := store.lockClaim(ctx, claim)
	if err != nil {
		return ExternalEffectPermit{}, err
	}
	if locked.cancelRequested.Valid {
		return ExternalEffectPermit{}, ErrJobNotControllable
	}
	startedAt, err := store.beginExternal(ctx, claim, locked.effectStarted.Valid)
	if err != nil {
		return ExternalEffectPermit{}, err
	}
	request, err := buildExternalEffectRequest(claim, spec, startedAt)
	if err != nil {
		return ExternalEffectPermit{}, err
	}
	digest, err := externalRequestDigest(request)
	if err != nil {
		return ExternalEffectPermit{}, err
	}
	var persistedAt time.Time
	err = store.database.QueryRow(
		ctx, markExternalAttemptEffectSQL,
		claim.attemptID, claim.jobID, claim.leaseOwnerID,
		claim.leaseGeneration, startedAt, spec.effectID, digest,
	).Scan(&persistedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ExternalEffectPermit{}, ErrJobLeaseLost
	}
	if err != nil || !persistedAt.Equal(startedAt) {
		return ExternalEffectPermit{}, ErrExternalEffectPermit
	}
	return newExternalEffectPermit(owner, claim, request, startedAt)
}

func (store *boundDatabaseJobStore) lockClaim(
	ctx context.Context,
	claim JobClaim,
) (lockedJobClaim, error) {
	arguments := append(authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration)
	var locked lockedJobClaim
	err := store.database.QueryRow(ctx, lockJobClaimSQL, arguments...).Scan(
		&locked.cancelRequested, &locked.effectStarted, &locked.leaseUntil,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return lockedJobClaim{}, ErrJobLeaseLost
	}
	return locked, err
}

func (store *boundDatabaseJobStore) beginExternal(
	ctx context.Context,
	claim JobClaim,
	alreadyStarted bool,
) (time.Time, error) {
	statement := markExternalJobEffectSQL
	if alreadyStarted {
		statement = observeRetriedExternalEffectSQL
	}
	arguments := append(authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration)
	var startedAt time.Time
	err := store.database.QueryRow(ctx, statement, arguments...).Scan(&startedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobLeaseLost
	}
	return startedAt, err
}

func (store *boundDatabaseJobStore) failBeforeExternal(
	ctx context.Context,
	claim JobClaim,
	failure JobFailure,
) (JobFailureResult, error) {
	decision, err := decideJobFailure(claim, failure)
	if err != nil {
		return JobFailureResult{}, err
	}
	locked, err := store.lockClaim(ctx, claim)
	if err != nil || locked.effectStarted.Valid {
		return JobFailureResult{}, ErrExternalEffectPermit
	}
	var occurredAt time.Time
	err = store.database.QueryRow(
		ctx, failJobAttemptSQL, claim.attemptID, claim.jobID,
		claim.leaseOwnerID, claim.leaseGeneration, decision.outcome,
		string(failure.errorClass), failure.code, failure.digest, nil, nil,
	).Scan(&occurredAt)
	if err != nil {
		return JobFailureResult{}, fmt.Errorf("fail external attempt before effect: %w", err)
	}
	arguments := append(authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
		string(decision.state), string(failure.errorClass), failure.code,
		failure.digest, failure.retryDelay.Microseconds(), occurredAt, nil)
	var persisted time.Time
	var retryAt pgtype.Timestamptz
	err = store.database.QueryRow(ctx, failJobSQL, arguments...).Scan(&persisted, &retryAt)
	result := JobFailureResult{state: decision.state, occurredAt: persisted}
	if retryAt.Valid {
		result.runAt = retryAt.Time
	}
	return result, err
}
