package worker

import (
	"time"

	"github.com/google/uuid"
)

type analyticsDashboardExportValues struct{}

func (analyticsDashboardExportValues) NewUUIDv7() (uuid.UUID, error) {
	return uuid.NewV7()
}

func (analyticsDashboardExportValues) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
