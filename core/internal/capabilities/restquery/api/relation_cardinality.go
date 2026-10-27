package api

import (
	"errors"
	"fmt"
	"io"
)

var errInvalidRelationCardinality = errors.New("REST Query relation cardinality is invalid")

type RelationCardinality uint8

const (
	RelationCardinalityToOne RelationCardinality = iota + 1
	RelationCardinalityToMany
)

func ParseRelationCardinality(value string) (RelationCardinality, bool) {
	switch value {
	case "to_one":
		return RelationCardinalityToOne, true
	case "to_many":
		return RelationCardinalityToMany, true
	default:
		return 0, false
	}
}

func (cardinality RelationCardinality) String() string {
	switch cardinality {
	case RelationCardinalityToOne:
		return "to_one"
	case RelationCardinalityToMany:
		return "to_many"
	default:
		return ""
	}
}

func (cardinality RelationCardinality) GoString() string { return cardinality.String() }

func (cardinality RelationCardinality) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, cardinality.String())
}

func (cardinality RelationCardinality) Valid() bool {
	return cardinality == RelationCardinalityToOne ||
		cardinality == RelationCardinalityToMany
}

func (cardinality RelationCardinality) Validate() error {
	if !cardinality.Valid() {
		return errInvalidRelationCardinality
	}
	return nil
}

func (RelationCardinality) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
