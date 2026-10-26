package systemtransfer

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

func resultDigest(outcome, ownerKind string, ownerID uuid.UUID,
	revision int64, code string) [sha256.Size]byte {
	document := fmt.Sprintf("iam.system-transfer.receipt-result.v1\x00%s\x00%s\x00%s\x00%d\x00%s",
		outcome, ownerKind, ownerID, revision, code)
	return sha256.Sum256([]byte(document))
}

func completedReceipt(row SourceRow, outcome, ownerKind string,
	ownerID uuid.UUID, revision int64) (RowReceipt, error) {
	return NewRowReceipt(row.Index, outcome, ownerKind, ownerID, revision, "",
		row.SHA256, resultDigest(outcome, ownerKind, ownerID, revision, ""))
}

func rejectedReceipt(row SourceRow, code string) (RowReceipt, error) {
	return NewRowReceipt(row.Index, "rejected", "", uuid.Nil, 0, code,
		row.SHA256, resultDigest("rejected", "", uuid.Nil, 0, code))
}

func skippedReceipt(row SourceRow) (RowReceipt, error) {
	return NewRowReceipt(row.Index, "skipped", "", uuid.Nil, 0, "",
		row.SHA256, resultDigest("skipped", "", uuid.Nil, 0, ""))
}
