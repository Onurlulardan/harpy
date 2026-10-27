package systemtransfer

import "context"

func (worker *Worker) pendingRows(
	ctx context.Context,
	attempt Attempt,
	rows []SourceRow,
) ([]SourceRow, int, error) {
	indices := make([]int64, len(rows))
	for index, row := range rows {
		indices[index] = row.Index
	}
	existing, err := worker.receipts.ExistingReceipts(
		ctx, attempt, indices,
	)
	if err != nil {
		return nil, 0, err
	}
	pending := make([]SourceRow, 0, len(rows))
	for _, row := range rows {
		receipt, found := existing[row.Index]
		if !found {
			pending = append(pending, row)
			continue
		}
		if !receipt.Valid() || receipt.SourceSHA256 != row.SHA256 {
			return nil, 0, ErrReceiptConflict
		}
	}
	return pending, len(rows) - len(pending), nil
}
