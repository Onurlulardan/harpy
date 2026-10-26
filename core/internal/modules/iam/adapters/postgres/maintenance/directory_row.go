package maintenance

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/google/uuid"
)

type directorySourceRow struct {
	id             uuid.UUID
	userID         uuid.UUID
	kind           string
	normalized     []byte
	normalizedSHA  []byte
	state          string
	sourceRevision int64
}

func (row *directorySourceRow) valid() bool {
	if row == nil || row.id == uuid.Nil || row.userID == uuid.Nil ||
		(row.kind != "email_v1" && row.kind != "portal_v1") ||
		!validDirectoryState(row.state) || row.sourceRevision < 1 ||
		len(row.normalized) < 3 || len(row.normalized) > 254 ||
		len(row.normalizedSHA) != sha256.Size {
		return false
	}
	digest := sha256.Sum256(row.normalized)
	return hmac.Equal(digest[:], row.normalizedSHA)
}

func (row *directorySourceRow) wipe() {
	if row == nil {
		return
	}
	clear(row.normalized)
	clear(row.normalizedSHA)
}

func validDirectoryState(value string) bool {
	return value == "pending" || value == "verified" || value == "active" ||
		value == "tombstoned" || value == "retired"
}
