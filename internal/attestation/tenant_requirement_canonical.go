package attestation

import (
	"encoding/json"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func MarshalCanonicalTenantRequirementCatalog(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode tenant requirement catalog: %w", err)
	}
	return canonicalizeBounded(encoded, postgresv1.MaximumMigrationManifestBytes)
}
