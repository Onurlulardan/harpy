package migration

import (
	"testing"
)

func TestEmbeddedReleaseRegistryAuthorizesEveryEmbeddedManifest(t *testing.T) {
	registry, err := LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	digests := registry.Digests()
	if len(digests) == 0 {
		t.Fatal("embedded migration manifest registry is empty")
	}
	for _, digest := range digests {
		artifact, err := registry.MigrationArtifact(digest)
		if err != nil {
			t.Fatal(err)
		}
		if artifact.DigestSHA256() != digest {
			t.Fatalf("embedded filename digest %s differs from artifact", digest)
		}
		if err := registry.Authorize(artifact); err != nil {
			t.Fatal(err)
		}
	}
}
