package documents

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	"github.com/google/uuid"
)

type platformClock struct{}

func (platformClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

type platformIDs struct{}

func (platformIDs) NewUUIDv7() (foundation.ID, error) {
	value, err := uuid.NewV7()
	return foundation.ID(value), err
}

var _ foundation.Clock = platformClock{}
var _ foundation.IDSource = platformIDs{}
