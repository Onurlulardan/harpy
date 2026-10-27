package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (repository *repository) MarkExternalEffectStarted(
	ctx context.Context,
	claim JobClaim,
	spec ExternalEffectRequestSpec,
) (ExternalEffectPermit, error) {
	if ctx == nil || repository == nil ||
		!repository.authority.matchesClaim(claim) {
		return ExternalEffectPermit{}, ErrAuthorityMismatch
	}
	if _, err := buildExternalEffectRequest(claim, spec, claim.startedAt); err != nil {
		return ExternalEffectPermit{}, err
	}
	locked, err := repository.lockClaim(ctx, claim)
	if err != nil {
		return ExternalEffectPermit{}, err
	}
	if locked.cancelRequested.Valid {
		return ExternalEffectPermit{}, ErrJobNotControllable
	}
	startedAt, err := repository.beginExternalEffect(
		ctx, claim, locked.effectStarted.Valid,
	)
	if err != nil {
		return ExternalEffectPermit{}, err
	}
	request, err := buildExternalEffectRequest(claim, spec, startedAt)
	if err != nil {
		return ExternalEffectPermit{}, err
	}
	digest, err := externalRequestDigest(request)
	if err != nil {
		return ExternalEffectPermit{}, fmt.Errorf("digest effect request: %w", err)
	}
	var persistedAt time.Time
	err = repository.database.QueryRow(
		ctx, markExternalAttemptEffectSQL,
		claim.attemptID, claim.jobID, claim.leaseOwnerID,
		claim.leaseGeneration, startedAt, spec.effectID, digest,
	).Scan(&persistedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ExternalEffectPermit{}, ErrJobLeaseLost
	}
	if err != nil {
		return ExternalEffectPermit{}, fmt.Errorf("mark bound attempt effect: %w", err)
	}
	if !persistedAt.Equal(startedAt) {
		return ExternalEffectPermit{}, ErrExternalEffectPermit
	}
	return newExternalEffectPermit(
		claim.authority.owner, claim, request, startedAt,
	)
}

func (repository *repository) beginExternalEffect(
	ctx context.Context,
	claim JobClaim,
	alreadyStarted bool,
) (time.Time, error) {
	arguments := append(
		authorityArguments(claim.authority),
		claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
	)
	statement := markExternalJobEffectSQL
	if alreadyStarted {
		statement = observeRetriedExternalEffectSQL
	}
	var startedAt time.Time
	err := repository.database.QueryRow(ctx, statement, arguments...).Scan(&startedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobLeaseLost
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("begin bound external effect: %w", err)
	}
	return startedAt, nil
}
