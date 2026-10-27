package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/fullrollout"
)

func RecordFullRollout(
	ctx context.Context,
	recorder fullrollout.Recorder,
	value fullrollout.PersistedCompletion,
) (fullrollout.Receipt, error) {
	return fullrollout.Record(ctx, recorder, value)
}
