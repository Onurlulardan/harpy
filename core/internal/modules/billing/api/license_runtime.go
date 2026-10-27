package billingapi

import (
	"context"
	"time"
)

type LicenseRuntimeState string

const (
	LicenseRuntimeValid    LicenseRuntimeState = "valid"
	LicenseRuntimeGrace    LicenseRuntimeState = "grace"
	LicenseRuntimeBuildOff LicenseRuntimeState = "build_off"
)

type LicenseRuntimeEntitlements struct {
	LicenseID        string
	EnvironmentID    string
	SnapshotSequence int64
	SnapshotSHA256   string
	State            LicenseRuntimeState
	Projection       LicenseEntitlementProjection
	DatabaseNow      time.Time
}

type LicenseRuntimeReader interface {
	ReadLicenseRuntimeEntitlements(
		context.Context,
		string,
	) (LicenseRuntimeEntitlements, error)
}
