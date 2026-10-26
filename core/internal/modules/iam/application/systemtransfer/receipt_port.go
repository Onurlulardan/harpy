package systemtransfer

import "context"

// ReceiptStore commits transfer receipts after canonical owner commands commit.
// AppendReceipts must be atomic and idempotent for the supplied transfer rows.
type ReceiptStore interface {
	ExistingReceipts(
		context.Context,
		Attempt,
		[]int64,
	) (map[int64]RowReceipt, error)
	AppendReceipts(
		context.Context,
		Attempt,
		[]RowReceipt,
		int64,
	) error
}
