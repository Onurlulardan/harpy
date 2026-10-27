package api

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

// NewJSONBFieldStorage seals one stable UUID key inside an owner-catalogued
// JSONB document column. Public field names and request values cannot become
// storage paths.
func NewJSONBFieldStorage(
	document sqlidentifier.Token,
	stableKey string,
) (FieldStorage, error) {
	key, valid := parseCatalogUUID(stableKey)
	if !valid {
		return FieldStorage{}, errInvalidFieldStorage
	}
	storage := FieldStorage{
		column: document, jsonbKey: key, hasJSONB: true, validated: true,
	}
	if err := storage.Validate(); err != nil {
		return FieldStorage{}, err
	}
	return storage, nil
}

func (storage FieldStorage) JSONBKey() (uuid.UUID, bool, error) {
	if err := storage.Validate(); err != nil {
		return uuid.Nil, false, err
	}
	return storage.jsonbKey, storage.hasJSONB, nil
}
