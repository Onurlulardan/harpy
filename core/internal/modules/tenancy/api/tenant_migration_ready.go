package api

import "time"

type TenantMigrationReady struct {
	ReleaseID               string
	EnvironmentID           string
	DeploymentMode          string
	RuntimeGeneration       int64
	DatabaseGeneration      int64
	MigrationManifestSHA256 string
	ReadyUntil              time.Time
}

func (ready TenantMigrationReady) IsValid() bool {
	return validTenantMigrationID(ready.ReleaseID) &&
		validTenantMigrationID(ready.EnvironmentID) &&
		(ready.DeploymentMode == "onpremise" || ready.DeploymentMode == "saas") &&
		ready.RuntimeGeneration > 0 && ready.DatabaseGeneration > 0 &&
		validOnboardingDigest(ready.MigrationManifestSHA256) &&
		!ready.ReadyUntil.IsZero()
}
