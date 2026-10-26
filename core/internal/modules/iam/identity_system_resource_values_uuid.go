package iam

import (
	"bytes"
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func (values identitySystemResourceValues) nullableUUID(
	fieldID string,
) (uuid.UUID, bool, error) {
	raw, found := values.raw(fieldID)
	if !found {
		return uuid.Nil, false, nil
	}
	if bytes.Equal(raw, []byte("null")) {
		return uuid.Nil, true, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return uuid.Nil, false, publicresource.ErrContract
	}
	id, err := uuid.Parse(text)
	if err != nil || id == uuid.Nil || id.String() != text {
		return uuid.Nil, false, publicresource.ErrContract
	}
	return id, true, nil
}

func (values identitySystemResourceValues) uuidList(
	fieldID string,
	required bool,
) ([]uuid.UUID, bool, error) {
	raw, found := values.raw(fieldID)
	if !found {
		if required {
			return nil, false, publicresource.ErrContract
		}
		return nil, false, nil
	}
	var texts []string
	if json.Unmarshal(raw, &texts) != nil || texts == nil || len(texts) > 100 {
		return nil, false, publicresource.ErrContract
	}
	result := make([]uuid.UUID, len(texts))
	seen := make(map[uuid.UUID]struct{}, len(texts))
	for index, text := range texts {
		id, err := uuid.Parse(text)
		if err != nil || id == uuid.Nil || id.String() != text {
			return nil, false, publicresource.ErrContract
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, false, publicresource.ErrContract
		}
		seen[id], result[index] = struct{}{}, id
	}
	return identitySortedIDs(result), true, nil
}
