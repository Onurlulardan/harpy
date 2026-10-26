package licenseauthority

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RuntimeGrant struct {
	LicenseID        uuid.UUID
	EnvironmentID    uuid.UUID
	SnapshotSequence int64
	SnapshotSHA256   string
	State            RuntimeState
	EntitlementKey   string
	ValueType        string
	BooleanValue     *bool
	IntegerValue     *int64
	Unlimited        bool
	Channel          string
	DatabaseNow      time.Time
}

type RuntimeStore interface {
	LoadLicenseRuntimeEntitlements(
		context.Context,
		uuid.UUID,
	) ([]RuntimeGrant, error)
}
