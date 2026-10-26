package schemacontract

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func LoadEmbedded(registry *migration.ReleaseRegistry) (EmbeddedContract, error) {
	return loadEmbedded(registry, binarycontract.LoadEmbedded)
}

func loadEmbedded(
	registry *migration.ReleaseRegistry,
	load func() (binarycontract.Artifact, error),
) (EmbeddedContract, error) {
	artifact, err := load()
	if err != nil {
		return EmbeddedContract{}, err
	}
	return authorizeEmbedded(registry, artifact)
}

func authorizeEmbedded(
	registry *migration.ReleaseRegistry,
	artifact binarycontract.Artifact,
) (EmbeddedContract, error) {
	if !artifact.IsVerified() {
		return EmbeddedContract{}, fmt.Errorf("embedded binary schema contract is not verified")
	}
	if err := registry.AuthorizeDigest(artifact.Value().MigrationManifestSHA256); err != nil {
		return EmbeddedContract{}, fmt.Errorf("authorize binary migration contract: %w", err)
	}
	return EmbeddedContract{artifact: artifact, authorized: true}, nil
}
