package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (worker *Worker) UpdateRecord(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence publicresource.ActorEvidence,
	attempt Attempt,
	row SourceRow,
	match IdentityMatch,
	values publicresource.FieldValues,
	key string,
) (RowReceipt, error) {
	input, err := publicresource.NewUpdateInput(
		attempt.Resource, match.ID, match.Revision, values, key,
	)
	if err != nil {
		return rejectedReceipt(row, "ROW_VALIDATION_FAILED")
	}
	result, err := worker.resources.Update(ctx, scope, evidence, input)
	if err != nil {
		return worker.commandFailure(row, err)
	}
	id, revision, err := mutationResultIdentity(result)
	if err != nil {
		return RowReceipt{}, err
	}
	return completedReceipt(row, "updated", "record", id, revision)
}
