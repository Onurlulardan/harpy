package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

const SystemExportMaximumAttempts = 3

type ExportJobClaim struct {
	TenantID, JobID, ExportID uuid.UUID
	AttemptID, LeaseOwnerID   uuid.UUID
	Resource                  publicresource.Resource
	AttemptNumber             int
	MaximumAttempts           int
	LeaseGeneration           int64
}

func (claim ExportJobClaim) Validate() error {
	if claim.TenantID == uuid.Nil || !exportV7(claim.JobID) ||
		!exportV7(claim.ExportID) || !exportV7(claim.AttemptID) ||
		!exportV7(claim.LeaseOwnerID) || !claim.Resource.Valid() ||
		claim.AttemptNumber < 1 || claim.AttemptNumber > SystemExportMaximumAttempts ||
		claim.MaximumAttempts != SystemExportMaximumAttempts || claim.LeaseGeneration < 1 {
		return ErrExportWorkerInvalid
	}
	return nil
}

type ExportJobLease interface {
	CheckpointExportLease(context.Context, ExportJobClaim) error
}

func exportV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7
}
