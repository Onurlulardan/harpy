package platformauthority

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/devjournal"
)

const (
	DevelopmentAckTrustFilename    = devjournal.DevelopmentAckTrustFilename
	DevelopmentReleaseRootFilename = devjournal.DevelopmentReleaseRootFilename
)

func ReconcileDevelopmentRegistryArtifacts(
	root string,
	ackSeedPath string,
	releaseSeedPath string,
	now time.Time,
) error {
	return devjournal.ReconcileDevelopmentRegistryArtifacts(
		root, ackSeedPath, releaseSeedPath, now,
	)
}

func PublishDevelopmentRecoveryTrust(
	developmentRoot string,
	recoveryRoot string,
) error {
	return devjournal.PublishDevelopmentRecoveryTrust(
		developmentRoot, recoveryRoot,
	)
}
