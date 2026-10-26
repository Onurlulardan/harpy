package api

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

var errInvalidResourceStorage = errors.New("REST Query resource storage is invalid")

type ResourceStorage struct {
	table     sqlidentifier.Token
	schema    sqlidentifier.Token
	qualified bool
	validated bool
}

func NewResourceStorage(table sqlidentifier.Token) (ResourceStorage, error) {
	storage := ResourceStorage{table: table, validated: true}
	if err := storage.Validate(); err != nil {
		return ResourceStorage{}, err
	}
	return storage, nil
}

func NewQualifiedResourceStorage(
	schema sqlidentifier.Token,
	table sqlidentifier.Token,
) (ResourceStorage, error) {
	if !sqlidentifier.IsStatic(schema) || !sqlidentifier.IsStatic(table) {
		return ResourceStorage{}, errInvalidResourceStorage
	}
	storage := ResourceStorage{
		table: table, schema: schema, qualified: true, validated: true,
	}
	if storage.Validate() != nil {
		return ResourceStorage{}, errInvalidResourceStorage
	}
	return storage, nil
}

func (storage ResourceStorage) Validate() error {
	if !storage.validated || storage.table.Validate() != nil ||
		storage.qualified && (!sqlidentifier.IsStatic(storage.schema) ||
			!sqlidentifier.IsStatic(storage.table)) ||
		!storage.qualified && storage.schema.Validate() == nil {
		return errInvalidResourceStorage
	}
	return nil
}

func (storage ResourceStorage) Schema() (sqlidentifier.Token, bool, error) {
	if storage.Validate() != nil {
		return sqlidentifier.Token{}, false, errInvalidResourceStorage
	}
	if !storage.qualified {
		return sqlidentifier.Token{}, false, nil
	}
	return storage.schema, true, nil
}

func (storage ResourceStorage) Table() (sqlidentifier.Token, error) {
	if err := storage.Validate(); err != nil {
		return sqlidentifier.Token{}, err
	}
	return storage.table, nil
}

func (storage ResourceStorage) FingerprintSHA256() ([sha256.Size]byte, error) {
	if err := storage.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	tableDigest, err := sqlidentifier.Fingerprint(storage.table)
	if err != nil || !storage.qualified {
		return tableDigest, err
	}
	schemaDigest, err := sqlidentifier.Fingerprint(storage.schema)
	if err != nil {
		return [sha256.Size]byte{}, err
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte("restquery_qualified_storage_v1\x00"))
	_, _ = digest.Write(schemaDigest[:])
	_, _ = digest.Write(tableDigest[:])
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result, nil
}

func (ResourceStorage) String() string   { return queryValueRedaction }
func (ResourceStorage) GoString() string { return queryValueRedaction }
func (ResourceStorage) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (ResourceStorage) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
