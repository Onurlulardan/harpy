package recordkernel

import (
	"bytes"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

// DecodeStoredRecord verifies the full field-id keyed storage shape without
// applying actor write policy. It is the only persistence rehydration entrypoint.
func DecodeStoredRecord(catalog Catalog, document []byte) (Record, []byte, error) {
	if catalog.SHA256() == [32]byte{} || len(document) < 2 || len(document) > MaximumRecordBytes {
		return Record{}, nil, ErrInputInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumRecordBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return Record{}, nil, ErrInputInvalid
	}
	var members map[string]json.RawMessage
	if json.Unmarshal(document, &members) != nil || members == nil ||
		len(members) != len(catalog.fields) {
		return Record{}, nil, ErrInputInvalid
	}
	values := make(map[uuid.UUID]Value, len(members))
	targets := 0
	for fieldIDText, raw := range members {
		fieldID, parseErr := uuid.Parse(fieldIDText)
		field, found := catalog.FieldByID(fieldID)
		if parseErr != nil || fieldID.String() != fieldIDText || !found {
			return Record{}, nil, ErrInputInvalid
		}
		value, count, decodeErr := decodeInputValue(field, raw)
		if decodeErr != nil {
			return Record{}, nil, decodeErr
		}
		targets += count
		if uint64(targets) > catalog.limits.RelationTargetsPerMutation {
			return Record{}, nil, ErrBudgetExceeded
		}
		values[fieldID] = value
	}
	record := NewRecord(values)
	normalized, err := canonicalFieldIDDocument(values)
	if err != nil || !bytes.Equal(normalized, document) ||
		!recordStorageNormalized(catalog, record) {
		return Record{}, nil, ErrInputInvalid
	}
	return record, append([]byte(nil), document...), nil
}
