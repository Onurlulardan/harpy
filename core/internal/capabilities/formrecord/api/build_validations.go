package recordkernel

import (
	"github.com/google/uuid"
)

func applyValidationRules(
	fields []Field,
	schema StorageSchema,
) ([]Field, error) {
	if len(schema.Validations) > MaximumCrossFieldRules {
		return nil, ErrCatalogInvalid
	}
	indexes := make(map[uuid.UUID]int, len(fields))
	for index, field := range fields {
		indexes[field.ID] = index
	}
	seen := map[string]struct{}{}
	for _, validation := range schema.Validations {
		if len(validation.FieldIDs) != 1 {
			return nil, ErrCatalogInvalid
		}
		fieldID, err := parseStableID(validation.FieldIDs[0])
		index, found := indexes[fieldID]
		key := validation.Kind + ":" + validation.FieldIDs[0]
		if err != nil || !found {
			return nil, ErrCatalogInvalid
		}
		if _, duplicate := seen[key]; duplicate {
			return nil, ErrCatalogInvalid
		}
		seen[key] = struct{}{}
		if applyValidation(&fields[index], validation) != nil {
			return nil, ErrCatalogInvalid
		}
	}
	for _, constraint := range schema.Constraints {
		if constraint.Kind == "unique" {
			continue
		}
		if len(constraint.FieldIDs) != 1 {
			return nil, ErrCatalogInvalid
		}
		fieldID, err := parseStableID(constraint.FieldIDs[0])
		index, found := indexes[fieldID]
		if err != nil || !found ||
			applyConstraint(&fields[index], constraint) != nil {
			return nil, ErrCatalogInvalid
		}
	}
	return fields, nil
}

func applyValidation(field *Field, validation ValidationDefinition) error {
	switch validation.Kind {
	case "required":
		if validation.Minimum != nil || validation.Maximum != nil {
			return ErrCatalogInvalid
		}
		field.Required = true
	case "immutable":
		if validation.Minimum != nil || validation.Maximum != nil {
			return ErrCatalogInvalid
		}
		field.Immutable = true
	case "text_length":
		return applyTextLength(field, validation.Minimum, validation.Maximum)
	default:
		return ErrCatalogInvalid
	}
	return nil
}

func applyConstraint(field *Field, constraint ConstraintDefinition) error {
	switch constraint.Kind {
	case "not_null":
		field.Required = true
		return nil
	case "text_length":
		return applyTextLength(field, constraint.Minimum, constraint.Maximum)
	default:
		return ErrCatalogInvalid
	}
}

func applyTextLength(field *Field, minimum, maximum *int) error {
	if field.TypeID != TypeText || minimum == nil || maximum == nil ||
		*minimum < 0 || *maximum < *minimum || *maximum > MaximumFieldBytes {
		return ErrCatalogInvalid
	}
	if field.TextMinimum != nil || field.TextMaximum != nil {
		if field.TextMinimum != nil && field.TextMaximum != nil &&
			*field.TextMinimum == *minimum && *field.TextMaximum == *maximum {
			return nil
		}
		return ErrCatalogInvalid
	}
	minCopy, maxCopy := *minimum, *maximum
	field.TextMinimum, field.TextMaximum = &minCopy, &maxCopy
	return nil
}
