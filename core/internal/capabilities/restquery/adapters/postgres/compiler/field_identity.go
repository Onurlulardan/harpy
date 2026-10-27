package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func fieldIdentity(
	field api.FieldDefinition,
) (uuid.UUID, api.LogicalType, error) {
	identifier, identifierErr := field.ID()
	fieldType, typeErr := field.Type()
	kind, kindErr := fieldType.Kind()
	if identifierErr != nil || typeErr != nil || kindErr != nil {
		return uuid.Nil, 0, ErrContract
	}
	return identifier, kind, nil
}
