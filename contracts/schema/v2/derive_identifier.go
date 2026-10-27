package schemav2

import (
	"encoding/base32"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type IdentifierKind string

const (
	IdentifierForm       IdentifierKind = "form"
	IdentifierField      IdentifierKind = "field"
	IdentifierRelation   IdentifierKind = "relation"
	IdentifierIndex      IdentifierKind = "index"
	IdentifierConstraint IdentifierKind = "constraint"
)

func DeriveIdentifier(kind IdentifierKind, logicalID string) (string, error) {
	prefix, supported := identifierPrefix(kind)
	parsed, err := uuid.Parse(logicalID)
	if !supported || err != nil || parsed == uuid.Nil || parsed.String() != logicalID ||
		parsed.Variant() != uuid.RFC4122 {
		return "", fmt.Errorf("schema logical identifier is invalid")
	}
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	encoded := strings.ToLower(encoding.EncodeToString(parsed[:]))
	return prefix + encoded, nil
}

func DeriveFieldIdentifier(
	logicalID string,
	generation int64,
	variant FieldVariant,
) (string, error) {
	if generation < 1 || generation > MaximumFieldGeneration {
		return "", fmt.Errorf("schema field generation is invalid")
	}
	suffix, supported := variantSuffix(variant)
	if !supported {
		return "", fmt.Errorf("schema field variant is invalid")
	}
	stem, err := DeriveIdentifier(IdentifierField, logicalID)
	if err != nil {
		return "", err
	}
	return stem + "_g" + strconv.FormatInt(generation, 36) + suffix, nil
}

func DeriveIndexIdentifier(logicalID string, generation int64) (string, error) {
	if generation < 1 || generation > MaximumFieldGeneration {
		return "", fmt.Errorf("schema index generation is invalid")
	}
	stem, err := DeriveIdentifier(IdentifierIndex, logicalID)
	if err != nil {
		return "", err
	}
	return stem + "_g" + strconv.FormatInt(generation, 36), nil
}

func DeriveConstraintIdentifier(logicalID string, generation int64) (string, error) {
	if generation < 1 || generation > MaximumFieldGeneration {
		return "", fmt.Errorf("schema constraint generation is invalid")
	}
	stem, err := DeriveIdentifier(IdentifierConstraint, logicalID)
	if err != nil {
		return "", err
	}
	return stem + "_g" + strconv.FormatInt(generation, 36), nil
}

func identifierPrefix(kind IdentifierKind) (string, bool) {
	switch kind {
	case IdentifierForm:
		return "frm_", true
	case IdentifierField:
		return "fld_", true
	case IdentifierRelation:
		return "rel_", true
	case IdentifierIndex:
		return "idx_", true
	case IdentifierConstraint:
		return "cst_", true
	default:
		return "", false
	}
}

func variantSuffix(variant FieldVariant) (string, bool) {
	switch variant {
	case VariantValue:
		return "", true
	case VariantFolded:
		return "_folded", true
	case VariantAccent:
		return "_accent", true
	default:
		return "", false
	}
}
