package catalog

import (
	"time"

	"github.com/google/uuid"
)

type runtimeValues struct{}

func (runtimeValues) New() (uuid.UUID, error) { return uuid.NewV7() }

func (runtimeValues) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
