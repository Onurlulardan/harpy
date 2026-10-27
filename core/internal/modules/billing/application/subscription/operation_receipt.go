package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type OperationReceipt struct {
	ID                uuid.UUID
	OperationID       string
	SourceOperationID uuid.UUID
	RequestSHA256     [32]byte
	ResultSHA256      [32]byte
	Result            SubscriptionSnapshot
	CompletedAt       time.Time
}

func (receipt OperationReceipt) valid() bool {
	return validV7(receipt.ID) && receipt.OperationID != "" &&
		validV7(receipt.SourceOperationID) && validDigest(receipt.RequestSHA256) &&
		validDigest(receipt.ResultSHA256) && receipt.Result.valid() &&
		receipt.ResultSHA256 == snapshotDigest(receipt.Result) &&
		validInstant(receipt.CompletedAt)
}

func newReceipt(
	id uuid.UUID,
	execution Execution,
	result SubscriptionSnapshot,
	completedAt time.Time,
) OperationReceipt {
	return OperationReceipt{
		ID: id, OperationID: execution.OperationID,
		SourceOperationID: execution.SourceOperationID,
		RequestSHA256:     execution.RequestSHA256,
		ResultSHA256:      snapshotDigest(result), Result: result, CompletedAt: completedAt,
	}
}

func receiptReplay(
	receipt *OperationReceipt,
	execution Execution,
) (CommitPlan, bool, error) {
	if receipt == nil {
		return CommitPlan{}, false, nil
	}
	if !receipt.valid() || receipt.OperationID != execution.OperationID ||
		receipt.SourceOperationID != execution.SourceOperationID ||
		receipt.RequestSHA256 != execution.RequestSHA256 {
		return CommitPlan{}, true, ErrIdempotencyReused
	}
	return CommitPlan{
		BeforeRevision: receipt.Result.Revision,
		After:          receipt.Result, Receipt: *receipt, Replay: true,
	}, true, nil
}
