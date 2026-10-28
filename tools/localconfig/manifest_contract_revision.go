package main

import "fmt"

// Bump when generated artifact contents or semantics change, even if names do not.
const currentArtifactContractRevision = 11

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
