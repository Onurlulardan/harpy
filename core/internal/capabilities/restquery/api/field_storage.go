package api

import (
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

var errInvalidFieldStorage = errors.New("REST Query field storage is invalid")

type FieldStorage struct {
	column    sqlidentifier.Token
	folded    sqlidentifier.Token
	accent    sqlidentifier.Token
	jsonbKey  uuid.UUID
	hasFolded bool
	hasAccent bool
	hasJSONB  bool
	validated bool
}

func NewFieldStorage(
	column sqlidentifier.Token,
	folded *sqlidentifier.Token,
	accent *sqlidentifier.Token,
) (FieldStorage, error) {
	storage := FieldStorage{column: column, validated: true}
	if folded != nil {
		storage.folded, storage.hasFolded = *folded, true
	}
	if accent != nil {
		storage.accent, storage.hasAccent = *accent, true
	}
	if err := storage.Validate(); err != nil {
		return FieldStorage{}, err
	}
	return storage, nil
}

func (storage FieldStorage) Validate() error {
	zero := sqlidentifier.Token{}
	if !storage.validated || storage.column.Validate() != nil ||
		storage.hasFolded != (storage.folded != zero) ||
		storage.hasAccent != (storage.accent != zero) ||
		storage.hasJSONB != (storage.jsonbKey != uuid.Nil) ||
		storage.hasAccent && !storage.hasFolded ||
		storage.hasJSONB && (storage.hasFolded || storage.hasAccent) ||
		storage.hasFolded && storage.folded == storage.column ||
		storage.hasAccent && (storage.accent == storage.column ||
			storage.accent == storage.folded) {
		return errInvalidFieldStorage
	}
	if storage.hasFolded && storage.folded.Validate() != nil {
		return errInvalidFieldStorage
	}
	if storage.hasAccent && storage.accent.Validate() != nil {
		return errInvalidFieldStorage
	}
	return nil
}

func (storage FieldStorage) Column() (sqlidentifier.Token, error) {
	if err := storage.Validate(); err != nil {
		return sqlidentifier.Token{}, err
	}
	return storage.column, nil
}

func (storage FieldStorage) FoldedCompanion() (sqlidentifier.Token, bool, error) {
	if err := storage.Validate(); err != nil {
		return sqlidentifier.Token{}, false, err
	}
	return storage.folded, storage.hasFolded, nil
}

func (storage FieldStorage) AccentCompanion() (sqlidentifier.Token, bool, error) {
	if err := storage.Validate(); err != nil {
		return sqlidentifier.Token{}, false, err
	}
	return storage.accent, storage.hasAccent, nil
}

func (FieldStorage) String() string   { return queryValueRedaction }
func (FieldStorage) GoString() string { return queryValueRedaction }
func (FieldStorage) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FieldStorage) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
