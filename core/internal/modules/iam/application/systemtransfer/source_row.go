package systemtransfer

import (
	"crypto/sha256"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
)

type SourceRow struct {
	Index  int64
	JSON   []byte
	SHA256 [sha256.Size]byte
}

func (row SourceRow) Validate() error {
	if row.Index < 1 || row.Index > 1_000_000 || len(row.JSON) < 2 ||
		int64(len(row.JSON)) > tabulartransfer.MaximumRowBytes ||
		row.SHA256 == [sha256.Size]byte{} || sha256.Sum256(row.JSON) != row.SHA256 {
		return ErrInvalidAttempt
	}
	return nil
}

func validateBatch(rows []SourceRow) error {
	if len(rows) < 1 || len(rows) > MaximumBatchRows {
		return ErrInvalidAttempt
	}
	seen := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		if row.Validate() != nil {
			return ErrInvalidAttempt
		}
		if _, duplicate := seen[row.Index]; duplicate {
			return ErrInvalidAttempt
		}
		seen[row.Index] = struct{}{}
	}
	return nil
}
