package api

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const maximumRelationIncludeRecords = 100

var errInvalidRelationDefinition = errors.New("REST Query relation definition is invalid")

type RelationDefinition struct {
	relationID             uuid.UUID
	publicName             string
	cardinality            RelationCardinality
	storage                RelationStorage
	target                 CatalogReference
	includeEnabled         bool
	maximumIncludedRecords int
	parsed                 bool
}

func NewRelationDefinition(
	relationID string,
	publicName string,
	cardinality RelationCardinality,
	storage RelationStorage,
	target CatalogReference,
	includeEnabled bool,
	maximumIncludedRecords int,
) (RelationDefinition, error) {
	identifier, identifierOK := parseCatalogUUID(relationID)
	definition := RelationDefinition{
		relationID: identifier, publicName: publicName, cardinality: cardinality,
		storage: storage, target: target, includeEnabled: includeEnabled,
		maximumIncludedRecords: maximumIncludedRecords, parsed: identifierOK,
	}
	if err := definition.Validate(); err != nil {
		return RelationDefinition{}, err
	}
	return definition, nil
}

func (definition RelationDefinition) Validate() error {
	if !definition.parsed || definition.relationID == uuid.Nil ||
		!validRelationPublicName(definition.publicName) ||
		definition.cardinality.Validate() != nil || definition.storage.Validate() != nil ||
		!definition.storage.matchesRelation(definition.relationID) ||
		definition.target.Validate() != nil ||
		!validRelationIncludePolicy(definition) {
		return errInvalidRelationDefinition
	}
	return nil
}

func (definition RelationDefinition) RelationID() (uuid.UUID, error) {
	if err := definition.Validate(); err != nil {
		return uuid.Nil, err
	}
	return definition.relationID, nil
}

func (definition RelationDefinition) PublicName() (string, error) {
	if err := definition.Validate(); err != nil {
		return "", err
	}
	return definition.publicName, nil
}

func (definition RelationDefinition) Cardinality() (RelationCardinality, error) {
	if err := definition.Validate(); err != nil {
		return 0, err
	}
	return definition.cardinality, nil
}

func (definition RelationDefinition) Storage() (RelationStorage, error) {
	if err := definition.Validate(); err != nil {
		return RelationStorage{}, err
	}
	return definition.storage, nil
}

func (definition RelationDefinition) Target() (CatalogReference, error) {
	if err := definition.Validate(); err != nil {
		return CatalogReference{}, err
	}
	return definition.target, nil
}

func (definition RelationDefinition) IncludeEnabled() (bool, error) {
	if err := definition.Validate(); err != nil {
		return false, err
	}
	return definition.includeEnabled, nil
}

func (definition RelationDefinition) MaximumIncludedRecords() (int, error) {
	if err := definition.Validate(); err != nil {
		return 0, err
	}
	return definition.maximumIncludedRecords, nil
}

func (RelationDefinition) String() string   { return queryValueRedaction }
func (RelationDefinition) GoString() string { return queryValueRedaction }
func (RelationDefinition) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RelationDefinition) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}

func validRelationPublicName(value string) bool {
	return value != "" && value != "_included" && utf8.ValidString(value) &&
		!strings.ContainsRune(value, '\x00') && norm.NFC.IsNormalString(value)
}

func validRelationIncludePolicy(definition RelationDefinition) bool {
	if !definition.includeEnabled {
		return definition.maximumIncludedRecords == 0
	}
	if definition.cardinality == RelationCardinalityToOne {
		return definition.maximumIncludedRecords == 1
	}
	return definition.cardinality == RelationCardinalityToMany &&
		definition.maximumIncludedRecords >= 1 &&
		definition.maximumIncludedRecords <= maximumRelationIncludeRecords
}
