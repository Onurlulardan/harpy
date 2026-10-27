package systemtransfer

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

type AttemptClaim struct {
	TenantID, ImportID, JobID uuid.UUID
	AttemptID, LeaseOwnerID   uuid.UUID
	LeaseGeneration           int64
	Resource                  publicresource.Resource
}

func (claim AttemptClaim) Validate() error {
	if claim.TenantID == uuid.Nil || claim.ImportID.Version() != 7 ||
		claim.JobID.Version() != 7 || claim.AttemptID.Version() != 7 ||
		claim.LeaseOwnerID.Version() != 7 || claim.LeaseGeneration < 1 ||
		claim.LeaseGeneration > 8 || !claim.Resource.Valid() {
		return ErrInvalidAttempt
	}
	return nil
}
