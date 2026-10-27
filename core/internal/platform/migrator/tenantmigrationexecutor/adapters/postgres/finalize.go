package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/tenantmigrationexecutor"
	"github.com/google/uuid"
)

func (storage *store) Complete(
	ctx context.Context, claim tenantmigrationexecutor.Claim,
	receiptID uuid.UUID, receipt tenantmigrationexecutor.MigrationReceipt,
) error {
	if err := storage.usable(ctx); err != nil {
		return err
	}
	var catalogSHA string
	var requirementCount int
	return storage.pool.QueryRow(ctx, completeSQL, claim.JobID, claim.Attempt,
		claim.ProfileInstanceID, claim.LeaseID, receiptID,
		receipt.ReceiptSHA256, receipt.EvidenceSHA256,
	).Scan(&catalogSHA, &requirementCount)
}

func (storage *store) Fail(
	ctx context.Context, claim tenantmigrationexecutor.Claim,
	failure tenantmigrationexecutor.Failure,
) error {
	if err := storage.usable(ctx); err != nil {
		return err
	}
	var recorded bool
	return storage.pool.QueryRow(ctx, failSQL, claim.JobID, claim.Attempt,
		claim.ProfileInstanceID, claim.LeaseID, failure.Outcome, failure.Code,
		failure.ReceiptID, failure.ReceiptSHA256, failure.EvidenceSHA256,
	).Scan(&recorded)
}
