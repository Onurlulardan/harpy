package application

import (
	"time"

	"github.com/google/uuid"
)

type IDSource interface {
	NewUUID() (uuid.UUID, error)
}

type Clock interface {
	Now() time.Time
}
