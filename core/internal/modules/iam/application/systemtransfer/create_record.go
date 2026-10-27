package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (worker *Worker) CreateRecord(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence publicresource.ActorEvidence,
	attempt Attempt,
	row SourceRow,
	values publicresource.FieldValues,
	key string,
) (RowReceipt, error) {
	input, err := publicresource.NewCreateInput(attempt.Resource, values, nil, key)
	if err != nil {
		return rejectedReceipt(row, "ROW_VALIDATION_FAILED")
	}
	prepared, err := worker.resources.PrepareCreate(ctx, input)
	if err != nil {
		return worker.commandFailure(row, err)
	}
	result, err := worker.resources.Create(ctx, scope, evidence, prepared)
	if err != nil {
		return worker.commandFailure(row, err)
	}
	id, revision, err := mutationResultIdentity(result)
	if err != nil {
		return RowReceipt{}, err
	}
	return completedReceipt(row, "created", "record", id, revision)
}
