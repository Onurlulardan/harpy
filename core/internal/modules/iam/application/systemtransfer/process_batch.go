package systemtransfer

import "context"

type BatchResult struct {
	Committed       int
	AlreadyComplete int
}

func (worker *Worker) ProcessBatch(
	ctx context.Context,
	attempt Attempt,
	rows []SourceRow,
) (BatchResult, error) {
	if ctx == nil || ctx.Err() != nil || worker == nil || attempt.Validate() != nil ||
		validateBatch(rows) != nil {
		return BatchResult{}, ErrInvalidAttempt
	}
	pending, complete, err := worker.pendingRows(ctx, attempt, rows)
	if err != nil {
		return BatchResult{}, err
	}
	committed := make([]RowReceipt, 0, len(pending))
	err = worker.scopes.InFreshImportBatch(ctx, attempt,
		func(scoped context.Context, scope BatchScope) error {
			evidence, authorizeErr := worker.AuthorizeBatch(scoped, scope, attempt)
			if authorizeErr != nil {
				return authorizeErr
			}
			var mappedBytes int64
			for _, row := range pending {
				receipt, size, rowErr := worker.processRow(
					scoped, scope.IAMScope(), evidence, attempt, row,
				)
				if rowErr != nil {
					return rowErr
				}
				mappedBytes += size
				if mappedBytes > MaximumMappedBatchByte {
					return ErrInvalidAttempt
				}
				committed = append(committed, receipt)
			}
			return nil
		})
	if err != nil {
		return BatchResult{}, err
	}
	if len(committed) > 0 {
		var sourceBytes int64
		for _, row := range pending {
			sourceBytes += int64(len(row.JSON))
		}
		if err = worker.receipts.AppendReceipts(
			ctx, attempt, committed, sourceBytes,
		); err != nil {
			return BatchResult{}, err
		}
	}
	return BatchResult{Committed: len(committed), AlreadyComplete: complete}, nil
}
