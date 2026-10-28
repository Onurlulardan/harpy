package migration

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type ReleaseRegistry struct {
	mu                sync.Mutex
	manifests         map[string][]byte
	known             map[string]struct{}
	digestList        []string
	documentContracts map[string]map[expectedDocument]struct{}
	freshPaths        []string
	readManifest      func(name string) ([]byte, error)
}

func (registry *ReleaseRegistry) Authorize(
	artifact attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest],
) error {
	if registry == nil || !artifact.IsVerified() {
		return fmt.Errorf("migration release registry or artifact is not verified")
	}
	expected, err := registry.canonicalBytes(artifact.DigestSHA256())
	if err != nil || !bytes.Equal(expected, artifact.CanonicalJSON()) {
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
	document, err := registry.canonicalBytes(digest)
	if err != nil {
		return attestation.VerifiedPostgresArtifact[postgresv1.MigrationManifest]{}, err
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
	if len(registry.digestList) > 0 {
		return append([]string(nil), registry.digestList...)
	}
	digests := make([]string, 0, len(registry.manifests))
	for digest := range registry.manifests {
		digests = append(digests, digest)
	}
	sort.Strings(digests)
	return digests
}
