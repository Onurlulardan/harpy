package tenantmigrationexecutor

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	"github.com/google/uuid"
)

// NewID creates the opaque identifiers used by the executor runtime.
func NewID() (uuid.UUID, error) {
	value, err := randomid.NewUUIDv4()
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(value)
}

// NewRelease validates the signed-release values before composition starts.
func NewRelease(
	id, deploymentMode string, runtimeGeneration int64, manifestSHA256 string,
) (Release, error) {
	releaseID, err := uuid.Parse(id)
	if err != nil {
		return Release{}, fmt.Errorf("tenant migration release id: %w", err)
	}
	release := Release{ID: releaseID, DeploymentMode: deploymentMode,
		RuntimeGeneration:       runtimeGeneration,
		MigrationManifestSHA256: manifestSHA256}
	if release.ID == uuid.Nil || release.RuntimeGeneration < 1 ||
		!validDigest(release.MigrationManifestSHA256) ||
		(release.DeploymentMode != "onpremise" && release.DeploymentMode != "saas") {
		return Release{}, fmt.Errorf("tenant migration release is invalid")
	}
	return release, nil
}
