package main

import (
	"fmt"

	"github.com/Onurlulardan/harpy/ops/local/artifactcontract"
)

const currentArtifactContractRevision = artifactcontract.CurrentRevision

func validateCurrentArtifactContractRevision(manifest setManifest) error {
	if manifest.ArtifactContractRevision != currentArtifactContractRevision {
		return fmt.Errorf("manifest artifact contract revision is not current")
	}
	return nil
}

func validateResettableArtifactContractRevision(manifest setManifest) error {
	if manifest.ArtifactContractRevision < 0 ||
		manifest.ArtifactContractRevision > currentArtifactContractRevision {
		return fmt.Errorf("manifest artifact contract revision is not resettable")
	}
	return nil
}
