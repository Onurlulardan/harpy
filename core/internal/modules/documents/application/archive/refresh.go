package archive

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (worker *Worker) refreshGate(ctx context.Context, plan Plan,
	jobID domain.ID,
) func(int) error {
	lastCount := -100
	lastAt := time.Time{}
	itemIDs := make([]domain.ID, len(plan.Entries))
	for index, entry := range plan.Entries {
		itemIDs[index] = entry.DocumentID
	}
	return func(completed int) error {
		now := worker.clock.Now()
		if completed != len(plan.Entries) && completed-lastCount < 100 &&
			!lastAt.IsZero() && now.Sub(lastAt) < 20*time.Second {
			return nil
		}
		if err := worker.jobs.Check(ctx, jobID); err != nil {
			return ErrCancelled
		}
		if err := worker.authorizer.Refresh(ctx, plan.TenantID, plan.ActorID,
			plan.AuthorizationRevision, itemIDs); err != nil {
			return ErrForbidden
		}
		lastCount, lastAt = completed, now
		return nil
	}
}
