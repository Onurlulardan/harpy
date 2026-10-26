package systemtransfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

func RowIdempotencyKey(transferID uuid.UUID, rowIndex int64,
	sourceDigest [32]byte) (string, error) {
	if transferID == uuid.Nil || transferID.Version() != 7 || rowIndex < 1 ||
		rowIndex > 1_000_000 || sourceDigest == [32]byte{} {
		return "", ErrInvalidReceipt
	}
	document := fmt.Sprintf("iam.system-transfer.row.v1\x00%s\x00%d\x00%s",
		transferID, rowIndex, hex.EncodeToString(sourceDigest[:]))
	digest := sha256.Sum256([]byte(document))
	return "iam-row-" + hex.EncodeToString(digest[:]), nil
}
