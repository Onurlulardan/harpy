package tenantmigration

import (
	"context"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type EnqueueBinding struct {
	JobID, OperationID      uuid.UUID
	ReleaseID               uuid.UUID
	RuntimeGeneration       int64
	DatabaseGeneration      int64
	MigrationManifestSHA256 string
}

type Repository interface {
	CurrentReady(context.Context) (tenantapi.TenantMigrationReady, bool, error)
	Enqueue(context.Context, EnqueueBinding) (tenantapi.TenantMigrationJob, error)
	Get(context.Context, uuid.UUID) (tenantapi.TenantMigrationJob, bool, error)
}

type NewID func() (uuid.UUID, error)

type Service struct {
	repository Repository
	newID      NewID
}

func NewService(repository Repository, newID NewID) (*Service, error) {
	if repository == nil || newID == nil {
		return nil, ErrUnavailable
	}
	return &Service{repository: repository, newID: newID}, nil
}
