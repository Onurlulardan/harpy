package analytics

import (
	"context"
	"fmt"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (effects *AsyncEffects) enqueueJob(
	ctx context.Context, writer analyticsAsyncWriter,
	authority platformtx.OwnerAuthority, kind, reference, dedupe string,
	runAt time.Time,
) error {
	job, err := effects.job(authority, kind, reference, dedupe, runAt)
	if err != nil {
		return err
	}
	result, err := writer.EnqueueJob(ctx, job)
	if err != nil || !result.Created || result.ID != job.JobID {
		if err != nil {
			return err
		}
		return fmt.Errorf("Analytics job enqueue result is invalid")
	}
	return nil
}
