package api

import (
	"time"

	auditmodule "github.com/Onurlulardan/harpy/core/internal/modules/audit"
	supportmodule "github.com/Onurlulardan/harpy/core/internal/modules/support"
	usagemodule "github.com/Onurlulardan/harpy/core/internal/modules/usage"
	"github.com/google/uuid"
)

type moduleRuntimeClock struct{}

func (moduleRuntimeClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

type supportRuntimeIDs struct{}

func (supportRuntimeIDs) NewV7() (supportmodule.RuntimeID, error) {
	value, err := uuid.NewV7()
	return supportmodule.RuntimeID(value), err
}

type usageRuntimeIDs struct{}

func (usageRuntimeIDs) NewV7() (uuid.UUID, error) { return uuid.NewV7() }

type auditRuntimeIDs struct{}

func (auditRuntimeIDs) NewUUIDv7() (uuid.UUID, error) { return uuid.NewV7() }

var _ supportmodule.RuntimeClock = moduleRuntimeClock{}
var _ supportmodule.RuntimeIDGenerator = supportRuntimeIDs{}
var _ usagemodule.Clock = moduleRuntimeClock{}
var _ usagemodule.IDGenerator = usageRuntimeIDs{}
var _ auditmodule.Clock = moduleRuntimeClock{}
var _ auditmodule.IDSource = auditRuntimeIDs{}
