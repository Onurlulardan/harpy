package billing

import (
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

type DocumentsUsage struct {
	catalogs   platformasync.WriteCatalogs
	nextID     usageapp.IDGenerator
	clock      usageapp.Clock
	jobActorID uuid.UUID
}

func NewDocumentsUsage(
	catalogs platformasync.WriteCatalogs,
	nextID usageapp.IDGenerator,
	clock func() time.Time,
	jobActorID uuid.UUID,
) (*DocumentsUsage, error) {
	_, _, catalogErr := catalogs.Snapshot()
	_, contractErr := usageapp.DocumentsStorageEntitlement()
	if contractErr != nil || catalogErr != nil ||
		nextID == nil || clock == nil || !documentsUsageV7(jobActorID) {
		return nil, usageapp.ErrConfiguration
	}
	return &DocumentsUsage{catalogs: catalogs, nextID: nextID, clock: clock,
		jobActorID: jobActorID}, nil
}

func documentsUsageV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}
