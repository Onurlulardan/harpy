package retention

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

const (
	JobKind         = "analytics.retention.purge"
	PayloadRevision = 1
	BatchLimit      = 100
)

func Reference(tenantID uuid.UUID) (string, string, error) {
	if tenantID == uuid.Nil {
		return "", "", ErrClaimInvalid
	}
	reference := "analytics-retention/" + tenantID.String()
	digest := sha256.Sum256([]byte(reference))
	return reference, hex.EncodeToString(digest[:]), nil
}
