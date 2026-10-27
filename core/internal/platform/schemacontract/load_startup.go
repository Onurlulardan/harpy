package schemacontract

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func LoadStartup(
	documents []migration.Document,
) (EmbeddedContract, *migration.ReleaseRegistry, error) {
	registry, schemaErr := migration.LoadStartupRegistry(documents)
	artifact, contractErr := binarycontract.LoadEmbedded()
	if contractErr != nil && schemaErr != nil {
		return EmbeddedContract{}, nil, fmt.Errorf("%w; %v", contractErr, schemaErr)
	}
	if contractErr != nil {
		return EmbeddedContract{}, nil, contractErr
	}
	if schemaErr != nil {
		return EmbeddedContract{}, nil, schemaErr
	}
	embedded, err := authorizeEmbedded(registry, artifact)
	if err != nil {
		return EmbeddedContract{}, nil, err
	}
	return embedded, registry, nil
}
