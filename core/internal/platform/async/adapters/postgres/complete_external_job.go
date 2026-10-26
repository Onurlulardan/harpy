package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) CompleteExternalJob(
	ctx context.Context,
	permit ExternalEffectPermit,
	receipt effectsv1.Receipt,
) (time.Time, error) {
	if !permit.valid(repository.identity) || receipt.Status != effectsv1.ReceiptSucceeded {
		return time.Time{}, ErrExternalEffectPermit
	}
	values, err := validateExternalReceipt(permit, receipt)
	if err != nil {
		return time.Time{}, err
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
			if err := storeExternalEffectReceipt(
				ctx, transaction, permit, receipt, values,
			); err != nil {
				return time.Time{}, err
			}
			completedAt, err := completeExternalAttempt(
				ctx, transaction, permit, values,
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
				ctx, completeExternalJobSQL, arguments...,
			).Scan(&persisted)
			if errors.Is(err, pgx.ErrNoRows) {
				return time.Time{}, ErrJobLeaseLost
			}
			if err != nil {
				return time.Time{}, fmt.Errorf("complete external job: %w", err)
			}
			return persisted, nil
		},
	)
}

func completeExternalAttempt(
	ctx context.Context,
	transaction pgx.Tx,
	permit ExternalEffectPermit,
	values externalReceiptValues,
) (time.Time, error) {
	claim := permit.claim
	var completedAt time.Time
	err := transaction.QueryRow(
		ctx, completeExternalJobAttemptSQL,
		claim.attemptID, claim.jobID, claim.leaseOwnerID,
		claim.leaseGeneration, permit.startedAt, values.receiptID,
		values.completedAt, values.effectID, values.requestSHA256,
	).Scan(&completedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, ErrJobLeaseLost
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("complete external job attempt: %w", err)
	}
	return completedAt, nil
}
