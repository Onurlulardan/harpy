package billingapi

import (
	"context"
	"time"

	licensev2 "github.com/Onurlulardan/harpy/contracts/license/v2"
)

type LicenseIdentity struct {
	LicenseID              string
	LicenseeOrganizationID string
	EnvironmentID          string
	DeliveryKind           licensev2.DeliveryKind
}

type LicenseVerificationBinding = LicenseIdentity

type LicenseLimit struct {
	Value     int64
	Unlimited bool
}

type LicenseEntitlementProjection struct {
	Boolean map[string]bool
	Limits  map[string]LicenseLimit
	Channel string
}

type VerifiedLicenseSnapshot interface {
	LicenseIdentity() LicenseIdentity
	Issuer() string
	Sequence() int64
	Digest() string
	Channel() string
	OfferingProjection() LicenseEntitlementProjection
}

type LicenseSnapshotChecker interface {
	CheckLicenseSnapshot(
		context.Context,
		[]byte,
		LicenseVerificationBinding,
		time.Time,
	) (VerifiedLicenseSnapshot, error)
}
