package migration

import (
	"bytes"
	"fmt"
	"sort"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type ReleaseRegistry struct {
	manifests         map[string][]byte
	documentContracts map[string]map[expectedDocument]struct{}
}

func (registry *ReleaseRegistry) Authorize(
	artifact attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest],
) error {
	if registry == nil || !artifact.IsVerified() {
		return fmt.Errorf("migration release registry or artifact is not verified")
	}
	expected, exists := registry.manifests[artifact.DigestSHA256()]
	if !exists || !bytes.Equal(expected, artifact.CanonicalJSON()) {
		return fmt.Errorf("migration manifest is not embedded in this binary")
	}
	return nil
}

func (registry *ReleaseRegistry) MigrationArtifact(
	digest string,
) (attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest], error) {
	if registry == nil {
		return attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest]{},
			fmt.Errorf("migration release registry is unavailable")
	}
	document, exists := registry.manifests[digest]
	if !exists {
		return attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest]{},
			fmt.Errorf("migration manifest is not embedded in this binary")
	}
	artifact, err := attestation.VerifyMigrationManifestArtifact(document)
	if err != nil {
		return attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest]{},
			fmt.Errorf("embedded migration manifest verification failed")
	}
	return artifact, nil
}

func (registry *ReleaseRegistry) Digests() []string {
	if registry == nil {
		return nil
	}
	digests := make([]string, 0, len(registry.manifests))
	for digest := range registry.manifests {
		digests = append(digests, digest)
	}
	sort.Strings(digests)
	return digests
}
