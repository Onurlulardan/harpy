package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type digestDocument struct {
	ContractVersion string          `json:"contract_version"`
	Resource        string          `json:"resource"`
	ResourceID      string          `json:"resource_id"`
	SchemaID        string          `json:"schema_id"`
	ProfileHash     string          `json:"profile_hash"`
	CoreFields      []coreFieldSpec `json:"core_fields"`
}

func finalDigest(
	resource Resource,
	schemaID string,
	profileHash [sha256.Size]byte,
	core []coreFieldSpec,
) ([sha256.Size]byte, error) {
	document := digestDocument{
		ContractVersion: "iam-rest-query-catalog.v2",
		Resource:        string(resource.Kind()), ResourceID: resource.ID().String(),
		SchemaID: schemaID, ProfileHash: hex.EncodeToString(profileHash[:]),
		CoreFields: core,
	}
	canonical, err := json.Marshal(document)
	if err != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	digest := sha256.Sum256(canonical)
	if digest == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, ErrContract
	}
	return digest, nil
}
