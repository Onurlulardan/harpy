package postgres

import (
	"context"
	"fmt"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

func (store *boundDatabaseJobStore) cancelOwnedJob(
	ctx context.Context,
	claim JobClaim,
	input platformasync.DatabaseJobCancellationInput,
) error {
	if input.JobID == claim.jobID {
		return platformasync.ErrDatabaseJobOutcomeInvalid
	}
	cancellation, err := NewOwnedJobCancellation(
		input.JobID, input.RequestedAt, input.Code, input.Digest,
	)
	if err != nil {
		return platformasync.ErrDatabaseJobOutcomeInvalid
	}
	repository := &repository{
		database: store.database,
		authority: authorityStamp{
			scopeID: claim.tenantID, groupID: claim.authority.groupID,
			groupGeneration: claim.authority.groupGeneration,
			domain:          claim.authority.domain, generation: claim.authority.generation,
		},
	}
	if _, err := repository.CancelOwnedJob(ctx, cancellation); err != nil {
		return fmt.Errorf("cancel database job sibling: %w", err)
	}
	return nil
}
