package billing

import (
	"time"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func NewConfiguredDocumentsUsage(catalogs platformasync.WriteCatalogs,
	jobActorID uuid.UUID,
) (*DocumentsUsage, error) {
	return NewDocumentsUsage(catalogs, uuid.NewV7, func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}, jobActorID)
}
