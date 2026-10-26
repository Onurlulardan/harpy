package formscleanup

import (
	"context"

	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *store) StageEnqueue(ctx context.Context,
	plan formscleanupapp.EnqueuePlan,
) (formscleanupapp.Candidate, error) {
	if !storage.usable(ctx) || !plan.OperationID.UUIDv7() ||
		!plan.ItemID.UUIDv7() || !plan.JobID.UUIDv7() || plan.RunAt.IsZero() {
		return formscleanupapp.Candidate{}, formscleanupapp.ErrInvalid
	}
	var itemID, jobID string
	var runAt dbInstant
	err := storage.queryRow(ctx, stageEnqueueSQL, plan.OperationID.String(),
		plan.ItemID.String(), plan.JobID.String(), plan.RunAt).
		Scan(&itemID, &jobID, &runAt)
	item, itemErr := domain.ParseID(itemID)
	job, jobErr := domain.ParseID(jobID)
	if err != nil || itemErr != nil || jobErr != nil ||
		!item.UUIDv7() || !job.UUIDv7() || !runAt.Valid {
		return formscleanupapp.Candidate{}, formscleanupapp.ErrUnavailable
	}
	return formscleanupapp.Candidate{
		ItemID: item, JobID: job, PurgeRunAt: runAt.Time.UTC(),
	}, nil
}
