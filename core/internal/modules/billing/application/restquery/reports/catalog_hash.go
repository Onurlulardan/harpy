package reports

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type catalogHashDocument struct {
	Version  string       `json:"version"`
	Resource resourceSpec `json:"resource"`
}

func catalogHash(spec resourceSpec) (api.CatalogHash, error) {
	document := catalogHashDocument{
		Version: "billing-platform-report-rest-query-catalog.v1", Resource: spec,
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		return api.CatalogHash{}, ErrContract
	}
	digest := sha256.Sum256(canonical)
	hash, err := api.NewCatalogHash(hex.EncodeToString(digest[:]))
	if err != nil {
		return api.CatalogHash{}, ErrContract
	}
	return hash, nil
}
