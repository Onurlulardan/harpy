package postgres

import (
	"encoding/json"

	"github.com/google/uuid"
)

type profileDocumentWire struct {
	SchemaID string          `json:"schema_id"`
	Values   json.RawMessage `json:"values"`
	SHA256   string          `json:"sha256"`
}

func decodeProfileDocument(
	document []byte,
) (profileDocumentWire, uuid.UUID, []byte, error) {
	var wire profileDocumentWire
	if decodeProfileJSON(document, &wire) != nil || len(wire.Values) == 0 {
		return profileDocumentWire{}, uuid.Nil, nil, errProfileSnapshotUnavailable
	}
	schemaID, err := decodeProfileUUID(wire.SchemaID)
	if err != nil {
		return profileDocumentWire{}, uuid.Nil, nil, errProfileSnapshotUnavailable
	}
	digest, err := decodeProfileSHA256(wire.SHA256)
	if err != nil {
		return profileDocumentWire{}, uuid.Nil, nil, errProfileSnapshotUnavailable
	}
	return wire, schemaID, digest, nil
}
