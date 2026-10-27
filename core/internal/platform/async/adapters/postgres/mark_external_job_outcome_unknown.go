package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) MarkExternalJobOutcomeUnknown(
	ctx context.Context,
	permit ExternalEffectPermit,
) (time.Time, error) {
	if !permit.valid(repository.identity) {
		return time.Time{}, ErrExternalEffectPermit
	}
	claim := permit.claim
	return withJobTransaction(
		ctx, repository, claim,
		func(transaction pgx.Tx) (time.Time, error) {
			locked, err := lockJobClaim(ctx, transaction, claim)
			if err != nil {
				return time.Time{}, err
			}
			if !locked.effectStarted.Valid ||
				locked.effectStarted.Time.After(permit.startedAt) {
				return time.Time{}, ErrExternalEffectPermit
			}
			requestDigest, err := externalRequestDigest(permit.request)
			if err != nil {
				return time.Time{}, ErrExternalEffectPermit
			}
			var receiptExists bool
			err = transaction.QueryRow(
				ctx, hasExternalEffectReceiptSQL,
				claim.authority.scopeKind, claim.authority.scopeID,
				claim.authority.groupID, claim.authority.groupGeneration,
				claim.authority.domain, claim.authority.generation,
				claim.effectKey, permit.request.EffectID, requestDigest,
			).Scan(&receiptExists)
			if err != nil {
				return time.Time{}, fmt.Errorf("inspect effect receipt certainty: %w", err)
			}
			if receiptExists {
				return time.Time{}, ErrExternalEffectReceipt
			}
			completedAt, err := markUnknownAttempt(
				ctx, transaction, permit, requestDigest,
			)
			if err != nil {
				return time.Time{}, err
			}
			arguments := append(
				authorityArguments(claim.authority),
				claim.jobID, claim.leaseOwnerID, claim.leaseGeneration,
				permit.startedAt, completedAt,
			)
			var persisted time.Time
			err = transaction.QueryRow(
				ctx, markExternalJobUnknownSQL, arguments...,
			).Scan(&persisted)
			if errors.Is(err, pgx.ErrNoRows) {
				return time.Time{}, ErrJobLeaseLost
			}
			if err != nil {
				return time.Time{}, fmt.Errorf("mark external job outcome unknown: %w", err)
			}
			return persisted, nil
		},
	)
}

func markUnknownAttempt(
	ctx context.Context,
	transaction pgx.Tx,
	permit ExternalEffectPermit,
	requestDigest string,
) (time.Time, error) {
	claim := permit.claim
	var completedAt time.Time
	err := transaction.QueryRow(
		ctx, markExternalAttemptUnknownSQL,
		claim.attemptID, claim.jobID, claim.leaseOwnerID,
		claim.leaseGeneration, permit.startedAt, jobsv1.UnknownOutcomeCode,
		permit.request.EffectID, requestDigest,
	).Scan(&completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobLeaseLost
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("mark external attempt outcome unknown: %w", err)
	}
	return completedAt, nil
}
