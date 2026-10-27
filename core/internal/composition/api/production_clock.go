package api

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/google/uuid"
)

type productionClock struct{}

func (productionClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func (clock productionClock) NewUUIDv7() (uuid.UUID, error) {
	return randomid.NewUUIDv7(clock.Now())
}
