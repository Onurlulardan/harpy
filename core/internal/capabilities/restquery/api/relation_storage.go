package api

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/physicalid"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

var errInvalidRelationStorage = errors.New("REST Query relation storage is invalid")

type RelationStorage struct {
	table      sqlidentifier.Token
	relationID uuid.UUID
	stateSeal  [sha256.Size]byte
	validated  bool
}

func NewRelationStorage(
	relationID string, table sqlidentifier.Token,
) (RelationStorage, error) {
	identifier, parsed := parseCatalogUUID(relationID)
	storage := RelationStorage{
		table: table, relationID: identifier, validated: parsed,
	}
	seal, err := relationStorageSeal(identifier, table)
	if err != nil {
		return RelationStorage{}, errInvalidRelationStorage
	}
	storage.stateSeal = seal
	if err := storage.Validate(); err != nil {
		return RelationStorage{}, err
	}
	return storage, nil
}

func (storage RelationStorage) Validate() error {
	expected, err := physicalid.Name("rel_", storage.relationID)
	seal, sealErr := relationStorageSeal(storage.relationID, storage.table)
	if !storage.validated || err != nil || sealErr != nil ||
		!sqlidentifier.Matches(storage.table, expected) ||
		storage.stateSeal == [sha256.Size]byte{} || storage.stateSeal != seal {
		return errInvalidRelationStorage
	}
	return nil
}

func (storage RelationStorage) Table() (sqlidentifier.Token, error) {
	if err := storage.Validate(); err != nil {
		return sqlidentifier.Token{}, err
	}
	return storage.table, nil
}

func (storage RelationStorage) FingerprintSHA256() ([sha256.Size]byte, error) {
	if err := storage.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return storage.stateSeal, nil
}

func (storage RelationStorage) matchesRelation(identifier uuid.UUID) bool {
	return storage.Validate() == nil && storage.relationID == identifier
}

func (RelationStorage) String() string   { return queryValueRedaction }
func (RelationStorage) GoString() string { return queryValueRedaction }
func (RelationStorage) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RelationStorage) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func relationStorageSeal(
	relationID uuid.UUID, table sqlidentifier.Token,
) ([sha256.Size]byte, error) {
	if relationID == uuid.Nil || relationID.Variant() != uuid.RFC4122 {
		return [sha256.Size]byte{}, errInvalidRelationStorage
	}
	tableFingerprint, err := sqlidentifier.Fingerprint(table)
	if err != nil {
		return [sha256.Size]byte{}, errInvalidRelationStorage
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte("restquery_relation_storage_v1\x00"))
	_, _ = digest.Write(relationID[:])
	_, _ = digest.Write(tableFingerprint[:])
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result, nil
}
