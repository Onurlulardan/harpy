package migration

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func (registry *ReleaseRegistry) canonicalBytes(digest string) ([]byte, error) {
	if registry == nil {
		return nil, fmt.Errorf("migration release registry is unavailable")
	}
	if registry.readManifest == nil {
		document, exists := registry.manifests[digest]
		if !exists || len(document) == 0 {
			return nil, fmt.Errorf("migration manifest is not embedded in this binary")
		}
		return document, nil
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if document, exists := registry.manifests[digest]; exists {
		return document, nil
	}
	if _, known := registry.known[digest]; !known {
		return nil, fmt.Errorf("migration manifest is not embedded in this binary")
	}
	raw, err := registry.readManifest(digest + releaseManifestSuffix)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(raw)
	if hex.EncodeToString(sum[:]) != digest {
		return nil, fmt.Errorf("migration manifest digest does not match its bytes")
	}
	artifact, err := attestation.VerifyMigrationManifestArtifact(raw)
	if err != nil || artifact.DigestSHA256() != digest {
		return nil, fmt.Errorf("embedded migration manifest verification failed")
	}
	canonical := artifact.CanonicalJSON()
	registry.manifests[digest] = canonical
	return canonical, nil
}
