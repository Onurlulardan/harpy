package postgres

import (
	"crypto/sha256"
	"encoding/json"

	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/retention"
	"github.com/google/uuid"
)

type digestDocument struct {
	JobID, AttemptID, TenantID uuid.UUID
	PayloadReference           string
	PayloadSHA256              string
	Outcome                    retentionapp.Outcome
}

func receiptDigest(jobID, attemptID, tenantID uuid.UUID,
	reference, payloadDigest string, outcome retentionapp.Outcome,
) [32]byte {
	document, _ := json.Marshal(digestDocument{JobID: jobID, AttemptID: attemptID,
		TenantID: tenantID, PayloadReference: reference,
		PayloadSHA256: payloadDigest, Outcome: outcome})
	return sha256.Sum256(document)
}
