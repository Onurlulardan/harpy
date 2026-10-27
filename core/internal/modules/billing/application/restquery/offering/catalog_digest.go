package offering

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type catalogDigestDocument struct {
	Version   string      `json:"version"`
	Resource  string      `json:"resource"`
	Schema    string      `json:"schema"`
	Table     string      `json:"table"`
	Fields    []fieldSpec `json:"fields"`
	StableID  string      `json:"stable_id"`
	SearchIDs []string    `json:"search_ids"`
}

func catalogHash() (api.CatalogHash, error) {
	document := catalogDigestDocument{
		Version:  "billing-offering-rest-query-catalog.v3",
		Resource: ResourceName, Schema: "platform", Table: "billing_offering_query_v2",
		Fields: append([]fieldSpec{}, offeringFields...), StableID: fieldID,
		SearchIDs: []string{fieldOfferingKey},
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
