package migrationmanifest

import (
	"fmt"

	"github.com/Onurlulardan/harpy/tools/internal/migrationstreamcatalog"
)

func loadMigrationStreams(
	repositoryRoot string,
	claims []StorageClaim,
) ([]MigrationStream, error) {
	manifest, err := migrationstreamcatalog.Load(repositoryRoot)
	if err != nil {
		return nil, fmt.Errorf("load migration stream contract: %w", err)
	}
	streams := manifest.Streams()
	if len(streams) == 0 {
		return nil, fmt.Errorf("migration streams contract is empty")
	}
	claimSet := make(map[string]bool, len(claims))
	coveredClaims := make(map[string]bool, len(claims))
	for _, claim := range claims {
		claimSet[claim.Owner+"|"+claim.Directory+"|"+claim.Scope] = true
	}
	for _, stream := range streams {
		claimKey := stream.Owner + "|" + stream.Directory + "|" + stream.Scope
		if !claimSet[claimKey] {
			return nil, fmt.Errorf("migration stream has no storage owner claim")
		}
		coveredClaims[claimKey] = true
	}
	if len(coveredClaims) != len(claimSet) {
		return nil, fmt.Errorf("storage migration claim has no migration stream")
	}
	return streams, nil
}
