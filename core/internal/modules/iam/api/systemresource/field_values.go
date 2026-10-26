package systemresource

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const (
	maximumFieldCount  = 500
	maximumValuesBytes = 1 << 20
)

type FieldValues struct {
	values    map[ID]json.RawMessage
	canonical []byte
	digest    [32]byte
	validated bool
}

func NewFieldValues(values map[ID]json.RawMessage) (FieldValues, error) {
	if values == nil || len(values) > maximumFieldCount {
		return FieldValues{}, ErrContract
	}
	members := make(map[string]json.RawMessage, len(values))
	for fieldID, value := range values {
		if !fieldID.Valid() || len(value) == 0 {
			return FieldValues{}, ErrContract
		}
		members[fieldID.String()] = append(json.RawMessage(nil), value...)
	}
	encoded, err := json.Marshal(members)
	if err != nil {
		return FieldValues{}, ErrContract
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, maximumValuesBytes)
	if err != nil {
		return FieldValues{}, ErrContract
	}
	owned, err := decodeFieldValues(canonical)
	if err != nil {
		return FieldValues{}, err
	}
	result := FieldValues{
		values: owned, canonical: canonical, digest: sha256.Sum256(canonical), validated: true,
	}
	if result.Validate() != nil {
		return FieldValues{}, ErrContract
	}
	return result, nil
}

func decodeFieldValues(document []byte) (map[ID]json.RawMessage, error) {
	var members map[string]json.RawMessage
	if json.Unmarshal(document, &members) != nil || members == nil ||
		len(members) > maximumFieldCount {
		return nil, ErrContract
	}
	result := make(map[ID]json.RawMessage, len(members))
	for key, value := range members {
		fieldID, err := ParseID(key)
		if err != nil {
			return nil, ErrContract
		}
		result[fieldID] = append(json.RawMessage(nil), value...)
	}
	return result, nil
}

func (values FieldValues) Validate() error {
	if !values.validated || values.values == nil ||
		len(values.values) > maximumFieldCount || len(values.canonical) < 2 ||
		len(values.canonical) > maximumValuesBytes || values.digest == [32]byte{} ||
		sha256.Sum256(values.canonical) != values.digest {
		return ErrContract
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		values.canonical, maximumValuesBytes,
	)
	if err != nil || !bytes.Equal(canonical, values.canonical) {
		return ErrContract
	}
	return nil
}

func (values FieldValues) Len() int {
	if values.Validate() != nil {
		return 0
	}
	return len(values.values)
}

func (values FieldValues) Values() (map[ID]json.RawMessage, error) {
	if values.Validate() != nil {
		return nil, ErrContract
	}
	result := make(map[ID]json.RawMessage, len(values.values))
	for fieldID, value := range values.values {
		result[fieldID] = append(json.RawMessage(nil), value...)
	}
	return result, nil
}

func (values FieldValues) CanonicalJSON() ([]byte, error) {
	if values.Validate() != nil {
		return nil, ErrContract
	}
	return append([]byte(nil), values.canonical...), nil
}

func (values FieldValues) SHA256() ([32]byte, error) {
	if values.Validate() != nil {
		return [32]byte{}, ErrContract
	}
	return values.digest, nil
}
