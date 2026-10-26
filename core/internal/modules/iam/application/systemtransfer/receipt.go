package systemtransfer

import (
	"regexp"

	"github.com/google/uuid"
)

var receiptCodePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)

type RowReceipt struct {
	RowIndex                      int64
	Outcome, OwnerKind, ErrorCode string
	OwnerID                       uuid.UUID
	OwnerRevision                 int64
	SourceSHA256, ResultSHA256    [32]byte
}

func NewRowReceipt(rowIndex int64, outcome, ownerKind string, ownerID uuid.UUID,
	ownerRevision int64, errorCode string, source, result [32]byte) (RowReceipt, error) {
	receipt := RowReceipt{RowIndex: rowIndex, Outcome: outcome, OwnerKind: ownerKind,
		OwnerID: ownerID, OwnerRevision: ownerRevision, ErrorCode: errorCode,
		SourceSHA256: source, ResultSHA256: result}
	if !receipt.Valid() {
		return RowReceipt{}, ErrInvalidReceipt
	}
	return receipt, nil
}

func (receipt RowReceipt) Valid() bool {
	if receipt.RowIndex < 1 || receipt.RowIndex > 1_000_000 ||
		receipt.SourceSHA256 == [32]byte{} || receipt.ResultSHA256 == [32]byte{} {
		return false
	}
	if receipt.Outcome == "skipped" {
		return receipt.OwnerKind == "" && receipt.OwnerID == uuid.Nil &&
			receipt.OwnerRevision == 0 && receipt.ErrorCode == ""
	}
	if receipt.Outcome == "rejected" {
		return receipt.OwnerKind == "" && receipt.OwnerID == uuid.Nil &&
			receipt.OwnerRevision == 0 && receiptCodePattern.MatchString(receipt.ErrorCode)
	}
	return (receipt.Outcome == "created" || receipt.Outcome == "updated") &&
		(receipt.OwnerKind == "record" || receipt.OwnerKind == "invitation") &&
		receipt.OwnerID != uuid.Nil && receipt.OwnerRevision > 0 && receipt.ErrorCode == ""
}
