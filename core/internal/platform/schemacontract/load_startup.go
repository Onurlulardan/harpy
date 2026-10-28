package schemacontract

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func LoadStartup(
	documents []migration.Document,
) (EmbeddedContract, *migration.ReleaseRegistry, error) {
	return loadStartup(documents, binarycontract.LoadEmbedded, migration.LoadReleaseRegistry)
}

func loadStartup(
	documents []migration.Document,
	loadContract func() (binarycontract.Artifact, error),
	loadRelease func() (*migration.ReleaseRegistry, error),
) (EmbeddedContract, *migration.ReleaseRegistry, error) {
	fresh, schemaErr := migration.LoadStartupRegistry(documents)
	artifact, contractErr := loadContract()
	if contractErr != nil && schemaErr != nil {
		return EmbeddedContract{}, nil, fmt.Errorf("%w; %v", contractErr, schemaErr)
	}
	if contractErr != nil {
		return EmbeddedContract{}, nil, contractErr
	}
	if schemaErr != nil {
		return EmbeddedContract{}, nil, schemaErr
	}
	if fresh.AuthorizeDigest(artifact.Value().MigrationManifestSHA256) == nil {
		embedded, err := authorizeEmbedded(fresh, artifact)
		if err != nil {
			return EmbeddedContract{}, nil, err
		}
		return embedded, fresh, nil
	}
	release, err := loadRelease()
	if err != nil {
		return EmbeddedContract{}, nil, fmt.Errorf("authorize binary migration contract: %w", err)
	}
	embedded, err := authorizeEmbedded(release, artifact)
	if err != nil {
		return EmbeddedContract{}, nil, err
	}
	return embedded, release, nil
}
