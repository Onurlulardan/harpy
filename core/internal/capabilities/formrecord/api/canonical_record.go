package recordkernel

import (
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

func canonicalFieldIDDocument(values map[uuid.UUID]Value) ([]byte, error) {
	members := make(map[string]json.RawMessage, len(values))
	for fieldID, value := range values {
		members[fieldID.String()] = value.CanonicalJSON()
	}
	encoded, err := json.Marshal(members)
	if err != nil {
		return nil, ErrInputInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, MaximumRecordBytes)
	if err != nil {
		return nil, ErrBudgetExceeded
	}
	return canonical, nil
}
