package api

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

var errInvalidFieldDefinition = errors.New("REST Query field definition is invalid")

type FieldDefinition struct {
	id           uuid.UUID
	publicName   string
	fieldType    FieldType
	storage      FieldStorage
	capabilities FieldCapabilities
	parsed       bool
}

func NewFieldDefinition(
	id string,
	publicName string,
	fieldType FieldType,
	storage FieldStorage,
	capabilities FieldCapabilities,
) (FieldDefinition, error) {
	identifier, identifierOK := parseCatalogUUID(id)
	definition := FieldDefinition{
		id: identifier, publicName: publicName, fieldType: fieldType.clone(),
		storage: storage, capabilities: capabilities.clone(), parsed: identifierOK,
	}
	if err := definition.Validate(); err != nil {
		return FieldDefinition{}, err
	}
	return definition, nil
}

func (definition FieldDefinition) Validate() error {
	if !definition.parsed || definition.id == uuid.Nil ||
		!validFieldPublicName(definition.publicName) ||
		definition.fieldType.Validate() != nil || definition.storage.Validate() != nil ||
		definition.capabilities.Validate() != nil ||
		!definition.capabilities.validFor(definition.fieldType) ||
		!fieldStorageMatchesType(definition.fieldType, definition.storage) {
		return errInvalidFieldDefinition
	}
	return nil
}

func (definition FieldDefinition) ID() (uuid.UUID, error) {
	if err := definition.Validate(); err != nil {
		return uuid.Nil, err
	}
	return definition.id, nil
}

func (definition FieldDefinition) PublicName() (string, error) {
	if err := definition.Validate(); err != nil {
		return "", err
	}
	return definition.publicName, nil
}

func (definition FieldDefinition) Type() (FieldType, error) {
	if err := definition.Validate(); err != nil {
		return FieldType{}, err
	}
	return definition.fieldType.clone(), nil
}

func (definition FieldDefinition) Storage() (FieldStorage, error) {
	if err := definition.Validate(); err != nil {
		return FieldStorage{}, err
	}
	return definition.storage, nil
}

func (definition FieldDefinition) Capabilities() (FieldCapabilities, error) {
	if err := definition.Validate(); err != nil {
		return FieldCapabilities{}, err
	}
	return definition.capabilities.clone(), nil
}

func (FieldDefinition) String() string   { return queryValueRedaction }
func (FieldDefinition) GoString() string { return queryValueRedaction }
func (FieldDefinition) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FieldDefinition) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func (definition FieldDefinition) clone() FieldDefinition {
	result := definition
	result.fieldType = definition.fieldType.clone()
	result.capabilities = definition.capabilities.clone()
	return result
}

func validFieldPublicName(value string) bool {
	return value != "" && value != "_included" && utf8.ValidString(value) &&
		!strings.ContainsRune(value, '\x00') && norm.NFC.IsNormalString(value)
}
