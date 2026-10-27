package profilepostgres

import (
	"encoding/hex"
	"encoding/json"

	"github.com/google/uuid"
)

func decodeProfileDocument(payload []byte) ([]byte, uuid.UUID, [32]byte, error) {
	var wire profileDocumentWire
	var digest [32]byte
	if len(payload) == 0 || json.Unmarshal(payload, &wire) != nil {
		return nil, uuid.Nil, digest, errProfile
	}
	schemaID, err := uuid.Parse(wire.SchemaID)
	decoded, digestErr := hex.DecodeString(wire.SHA256)
	if err != nil || digestErr != nil || len(decoded) != len(digest) || len(wire.Values) < 2 {
		return nil, uuid.Nil, digest, errProfile
	}
	copy(digest[:], decoded)
	return append([]byte(nil), wire.Values...), schemaID, digest, nil
}
