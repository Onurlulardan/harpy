package maintenance

import (
	"encoding/hex"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
)

func claimValues(claim maintenanceapp.Claim) ([]any, error) {
	if claim.Validate() != nil {
		return nil, maintenanceapp.ErrClaimInvalid
	}
	digest, err := hex.DecodeString(claim.PayloadSHA256)
	if err != nil || len(digest) != 32 {
		return nil, maintenanceapp.ErrClaimInvalid
	}
	return []any{
		claim.TenantID, claim.JobID, claim.AttemptID, claim.PayloadRevision,
		claim.PayloadReference, digest,
	}, nil
}
