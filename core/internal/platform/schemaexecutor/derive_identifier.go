package schemaexecutor

import (
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/google/uuid"
)

const maximumFieldGeneration = schemav1.MaximumFieldGeneration

func canonicalUUID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, fmt.Errorf("UUID is not canonical and non-zero")
	}
	return parsed, nil
}

func deriveObjectIdentifier(prefix, logicalID string) (string, error) {
	kind := map[string]schemav1.IdentifierKind{
		"frm_": schemav1.IdentifierForm, "fld_": schemav1.IdentifierField,
		"rel_": schemav1.IdentifierRelation, "idx_": schemav1.IdentifierIndex,
		"cst_": schemav1.IdentifierConstraint,
	}[prefix]
	if kind == "" {
		return "", fmt.Errorf("identifier prefix is outside contract")
	}
	return schemav1.DeriveIdentifier(kind, logicalID)
}

func deriveFieldIdentifier(logicalID string, generation int64) (string, error) {
	return schemav1.DeriveFieldIdentifier(logicalID, generation)
}
