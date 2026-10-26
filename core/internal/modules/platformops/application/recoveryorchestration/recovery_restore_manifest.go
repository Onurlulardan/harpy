package recoveryorchestration

import (
	"fmt"
	"time"

	environmentrecoveryv1 "github.com/Onurlulardan/harpy/contracts/environmentrecovery/v1"
	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

func parseRecoveryManifest(
	document []byte,
	environmentID, recoverySetID string,
	trust recoverysigningv1.Bundle,
) (environmentrecoveryv1.Artifact, recoveryset.ManifestPayload, error) {
	now := time.Now().UTC().Truncate(time.Microsecond)
	artifact, err := environmentrecoveryv1.Parse(document, trust, now)
	if err != nil ||
		artifact.Purpose() != environmentrecoveryv1.PurposeRecoveryManifest ||
		artifact.EnvironmentID() != environmentID ||
		artifact.ArtifactID() != recoverySetID {
		return environmentrecoveryv1.Artifact{},
			recoveryset.ManifestPayload{},
			fmt.Errorf("recovery manifest artifact is invalid")
	}
	payload, err := recoveryset.ParseManifest(artifact.PayloadBytes())
	if err != nil || payload.Identity.EnvironmentID != environmentID ||
		payload.RecoveryTrustBundleSHA256 != trust.SHA256() ||
		payload.RecoverySigningGeneration != trust.Generation() {
		return environmentrecoveryv1.Artifact{},
			recoveryset.ManifestPayload{},
			fmt.Errorf("recovery manifest payload is invalid")
	}
	return artifact, payload, nil
}
