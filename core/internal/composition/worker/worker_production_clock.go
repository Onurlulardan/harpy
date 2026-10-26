package worker

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/google/uuid"
)

type workerProductionClock struct{}

func (workerProductionClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func (clock workerProductionClock) NewUUIDv7() (uuid.UUID, error) {
	return randomid.NewUUIDv7(clock.Now())
}
