package postgres

import (
	"encoding/hex"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
)

type fieldChangeDocument struct {
	Field          string                `json:"field"`
	Classification domain.Classification `json:"classification"`
	Before         *string               `json:"before,omitempty"`
	After          *string               `json:"after,omitempty"`
	BeforeSHA256   string                `json:"before_sha256,omitempty"`
	AfterSHA256    string                `json:"after_sha256,omitempty"`
}

func encodeFieldChanges(changes []domain.FieldChange) ([]byte, error) {
	document := make([]fieldChangeDocument, 0, len(changes))
	for _, change := range changes {
		if change.Validate() != nil {
			return nil, application.ErrInvalid
		}
		item := fieldChangeDocument{Field: change.Field, Classification: change.Classification,
			Before: change.Before, After: change.After}
		if change.BeforeSHA256 != [32]byte{} {
			item.BeforeSHA256 = hex.EncodeToString(change.BeforeSHA256[:])
		}
		if change.AfterSHA256 != [32]byte{} {
			item.AfterSHA256 = hex.EncodeToString(change.AfterSHA256[:])
		}
		document = append(document, item)
	}
	return json.Marshal(document)
}

func decodeFieldChanges(document []byte) ([]domain.FieldChange, error) {
	var wire []fieldChangeDocument
	if len(document) == 0 || json.Unmarshal(document, &wire) != nil || len(wire) > 100 {
		return nil, application.ErrUnavailable
	}
	changes := make([]domain.FieldChange, 0, len(wire))
	for _, item := range wire {
		change := domain.FieldChange{Field: item.Field, Classification: item.Classification,
			Before: item.Before, After: item.After}
		if decodeDigest(item.BeforeSHA256, &change.BeforeSHA256) != nil ||
			decodeDigest(item.AfterSHA256, &change.AfterSHA256) != nil || change.Validate() != nil {
			return nil, application.ErrUnavailable
		}
		changes = append(changes, change)
	}
	return changes, nil
}

func decodeDigest(value string, target *[32]byte) error {
	if value == "" {
		return nil
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return application.ErrUnavailable
	}
	copy(target[:], decoded)
	return nil
}
