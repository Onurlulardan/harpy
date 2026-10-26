package postgres

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/retention"
)

func (repository *retentionRepository) RunRetentionCycle(
	ctx context.Context,
) (retention.CycleResult, error) {
	if ctx == nil || !repository.valid() {
		return retention.CycleResult{}, ErrRetentionStoreUnavailable
	}
	var acquired bool
	var counts retention.StepCounts
	err := repository.database.QueryRow(ctx, runRetentionCycleSQL).Scan(
		&acquired,
		&counts.JobsRedacted,
		&counts.OutboxRedacted,
		&counts.DeadLettersRedacted,
		&counts.InboxPurged,
		&counts.IdempotencyPurged,
		&counts.JobsPurged,
		&counts.DeadLettersPurged,
	)
	if err != nil {
		return retention.CycleResult{}, fmt.Errorf("run async retention cycle: %w", err)
	}
	result, err := retention.NewCycleResult(acquired, counts)
	if err != nil {
		return retention.CycleResult{}, fmt.Errorf("%w: %v", ErrRetentionResultInvalid, err)
	}
	return result, nil
}
