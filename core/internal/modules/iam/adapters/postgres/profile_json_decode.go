package postgres

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

func decodeProfileJSON(document []byte, destination any) error {
	if len(document) == 0 || destination == nil {
		return errProfileSnapshotUnavailable
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errProfileSnapshotUnavailable
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errProfileSnapshotUnavailable
	}
	return nil
}

func decodeProfileUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil || id.Variant() != uuid.RFC4122 ||
		id.String() != value {
		return uuid.Nil, errProfileSnapshotUnavailable
	}
	return id, nil
}

func decodeProfileSHA256(value string) ([]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != value {
		return nil, errProfileSnapshotUnavailable
	}
	return decoded, nil
}
