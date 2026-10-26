//go:build integration

package schemacontract

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func AuthorizeIntegrationArtifact(
	registry *migration.ReleaseRegistry,
	artifact binarycontract.Artifact,
) (EmbeddedContract, error) {
	return authorizeEmbedded(registry, artifact)
}
