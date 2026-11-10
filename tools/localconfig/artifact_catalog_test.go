package main

import "testing"

const expectedGeneratedArtifactCount = 100

func TestArtifactCatalogAndManifestHaveExactCount(t *testing.T) {
	set, now := generatedSetForTest(t)
	if got := len(expectedArtifactNames()); got != expectedGeneratedArtifactCount {
		t.Fatalf("artifact catalog count = %d, want %d", got, expectedGeneratedArtifactCount)
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(manifest.Files); got != expectedGeneratedArtifactCount {
		t.Fatalf("manifest file count = %d, want %d", got, expectedGeneratedArtifactCount)
	}
	if manifest.ArtifactContractRevision != currentArtifactContractRevision {
		t.Fatalf("artifact contract revision = %d", manifest.ArtifactContractRevision)
	}
}
