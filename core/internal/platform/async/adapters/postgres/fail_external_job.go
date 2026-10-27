package postgres

import (
	"context"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/jackc/pgx/v5"
)

func (repository *workerRepository) FailExternalJob(
	ctx context.Context,
	permit ExternalEffectPermit,
	receipt effectsv1.Receipt,
	failure JobFailure,
) (JobFailureResult, error) {
	if !permit.valid(repository.identity) || receipt.Status != effectsv1.ReceiptFailed ||
		receipt.ErrorClass == nil || string(*receipt.ErrorClass) != string(failure.errorClass) ||
		receipt.Code != failure.code {
		return JobFailureResult{}, ErrExternalEffectReceipt
	}
	values, err := validateExternalReceipt(permit, receipt)
	if err != nil {
		return JobFailureResult{}, err
	}
	claim := permit.claim
	decision, err := decideJobFailure(claim, failure)
	if err != nil {
		return JobFailureResult{}, err
	}
	return withJobTransaction(
		ctx, repository, claim,
		func(transaction pgx.Tx) (JobFailureResult, error) {
			locked, err := lockJobClaim(ctx, transaction, claim)
			if err != nil {
				return JobFailureResult{}, err
			}
			if !locked.effectStarted.Valid ||
				locked.effectStarted.Time.After(permit.startedAt) {
				return JobFailureResult{}, ErrExternalEffectPermit
			}
			if err := storeExternalEffectReceipt(
				ctx, transaction, permit, receipt, values,
			); err != nil {
				return JobFailureResult{}, err
			}
			occurredAt, err := failJobAttempt(
				ctx, transaction, claim, failure, decision,
				values.receiptID, permit.startedAt,
			)
			if err != nil {
				return JobFailureResult{}, err
			}
			return persistFailedJob(
				ctx, transaction, claim, failure, decision,
				occurredAt, permit.startedAt,
			)
		},
	)
}
