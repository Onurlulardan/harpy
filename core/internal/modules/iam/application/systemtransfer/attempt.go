package systemtransfer

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/tabularmapping/api"
	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

const (
	MaximumBatchRows       = tabulartransfer.MaximumPageRows
	MaximumMappedBatchByte = tabulartransfer.MaximumMappedPageBytes
)

type DuplicateAction string

const (
	DuplicateSkip   DuplicateAction = "skip"
	DuplicateUpdate DuplicateAction = "update"
)

func (action DuplicateAction) Valid() bool {
	return action == DuplicateSkip || action == DuplicateUpdate
}

type Attempt struct {
	ID, JobID                             uuid.UUID
	AttemptID                             uuid.UUID
	LeaseOwnerID                          uuid.UUID
	LeaseGeneration                       int64
	TenantID                              uuid.UUID
	ActorID                               uuid.UUID
	Resource                              publicresource.Resource
	AuthorizationVersion                  int64
	CatalogVersionID                      uuid.UUID
	CatalogSHA256, IdentityManifestSHA256 [32]byte
	DuplicateAction                       DuplicateAction
	Mapping                               tabularmapping.Program
}

func (attempt Attempt) Validate() error {
	if attempt.ID == uuid.Nil || attempt.ID.Version() != 7 ||
		attempt.JobID == uuid.Nil || attempt.JobID.Version() != 7 ||
		attempt.AttemptID == uuid.Nil || attempt.AttemptID.Version() != 7 ||
		attempt.LeaseOwnerID == uuid.Nil || attempt.LeaseOwnerID.Version() != 7 ||
		attempt.LeaseGeneration < 1 || attempt.LeaseGeneration > 8 ||
		attempt.TenantID == uuid.Nil || attempt.ActorID == uuid.Nil ||
		!attempt.Resource.Valid() || attempt.AuthorizationVersion < 1 ||
		attempt.CatalogVersionID == uuid.Nil ||
		attempt.CatalogVersionID.Version() != 7 ||
		attempt.CatalogSHA256 == [32]byte{} ||
		attempt.IdentityManifestSHA256 == [32]byte{} ||
		!attempt.DuplicateAction.Valid() ||
		attempt.Mapping.Contract() != tabularmapping.ContractV2 ||
		len(attempt.Mapping.Targets()) == 0 {
		return ErrInvalidAttempt
	}
	return nil
}
