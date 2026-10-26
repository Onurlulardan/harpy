package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (worker *Worker) processRow(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence publicresource.ActorEvidence,
	attempt Attempt,
	row SourceRow,
) (RowReceipt, int64, error) {
	mapped, err := attempt.Mapping.Apply(row.JSON)
	if err != nil {
		receipt, receiptErr := rejectedReceipt(row, "MAPPING_INVALID")
		return receipt, 0, receiptErr
	}
	values, err := decodeFieldValues(mapped)
	if err != nil {
		receipt, receiptErr := rejectedReceipt(row, "ROW_VALIDATION_FAILED")
		return receipt, int64(len(mapped)), receiptErr
	}
	manifest, err := ManifestFor(attempt.Resource)
	if err != nil {
		return RowReceipt{}, 0, err
	}
	match, err := worker.identities.FindIdentity(
		ctx, scope, evidence, manifest, values,
	)
	if err != nil || match.Validate() != nil {
		return RowReceipt{}, 0, publicresource.ErrUnavailable
	}
	key, err := RowIdempotencyKey(attempt.ID, row.Index, row.SHA256)
	if err != nil {
		return RowReceipt{}, 0, err
	}
	if match.Found && attempt.DuplicateAction == DuplicateSkip {
		receipt, receiptErr := skippedReceipt(row)
		return receipt, int64(len(mapped)), receiptErr
	}
	var receipt RowReceipt
	if match.Found {
		receipt, err = worker.UpdateRecord(
			ctx, scope, evidence, attempt, row, match, values, key,
		)
	} else if attempt.Resource == publicresource.Users {
		receipt, err = worker.inviteUser(ctx, scope, evidence, row, values, key)
	} else {
		receipt, err = worker.CreateRecord(
			ctx, scope, evidence, attempt, row, values, key,
		)
	}
	return receipt, int64(len(mapped)), err
}
