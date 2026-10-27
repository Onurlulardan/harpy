package schemaexecutor

import (
	"fmt"
	"strings"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
)

func validateIdentifierMappingV2(
	mapping schemav2.IdentifierMapping,
	resourceID string,
) error {
	if mapping.RelatedForms == nil || mapping.Fields == nil ||
		mapping.Relations == nil || mapping.Indexes == nil || mapping.Constraints == nil {
		return fmt.Errorf("identifier mapping arrays must be explicit")
	}
	if err := validateIdentifierV2(mapping.Form, schemav2.IdentifierForm, resourceID); err != nil {
		return fmt.Errorf("form mapping: %w", err)
	}
	if len(mapping.Fields) > 1200 || len(mapping.Relations) > 100 ||
		len(mapping.Indexes) > 128 || len(mapping.Constraints) > 256 {
		return fmt.Errorf("identifier mapping exceeds object budget")
	}
	seenNames := map[string]struct{}{mapping.Form.PhysicalName: {}}
	if err := validateIdentifiersV2(
		mapping.RelatedForms, schemav2.IdentifierForm, seenNames,
	); err != nil {
		return fmt.Errorf("related form mapping: %w", err)
	}
	if err := validateFieldsV2(mapping.Fields, seenNames); err != nil {
		return err
	}
	groups := []struct {
		label string
		kind  schemav2.IdentifierKind
		items []schemav2.Identifier
	}{
		{"relation", schemav2.IdentifierRelation, mapping.Relations},
	}
	for _, group := range groups {
		if err := validateIdentifiersV2(group.items, group.kind, seenNames); err != nil {
			return fmt.Errorf("%s mapping: %w", group.label, err)
		}
	}
	if err := validateConstraintIdentifiersV2(mapping.Constraints, seenNames); err != nil {
		return err
	}
	return validateIndexIdentifiersV2(mapping.Indexes, seenNames)
}

func validateIdentifierV2(
	value schemav2.Identifier,
	kind schemav2.IdentifierKind,
	expectedID string,
) error {
	if expectedID != "" && value.LogicalID != expectedID {
		return fmt.Errorf("logical ID differs from resource")
	}
	if value.Generation != 0 {
		return fmt.Errorf("unversioned identifier carries a generation")
	}
	expected, err := schemav2.DeriveIdentifier(kind, value.LogicalID)
	if err != nil || value.PhysicalName != expected {
		return fmt.Errorf("physical identifier is not deterministic")
	}
	return nil
}

func validateIndexIdentifiersV2(
	values []schemav2.Identifier, seenNames map[string]struct{},
) error {
	previous := ""
	for _, value := range values {
		expected, err := schemav2.DeriveIndexIdentifier(value.LogicalID, value.Generation)
		key := value.LogicalID + ":" + fmt.Sprintf("%010d", value.Generation)
		if err != nil || value.PhysicalName != expected {
			return fmt.Errorf("index physical identifier is not deterministic")
		}
		if previous != "" && previous >= key {
			return fmt.Errorf("index identifiers are not sorted and unique")
		}
		if _, exists := seenNames[value.PhysicalName]; exists {
			return fmt.Errorf("physical identifier is duplicated")
		}
		seenNames[value.PhysicalName] = struct{}{}
		previous = key
	}
	return nil
}

func validateIdentifiersV2(
	values []schemav2.Identifier,
	kind schemav2.IdentifierKind,
	seenNames map[string]struct{},
) error {
	previous := ""
	for _, value := range values {
		if previous != "" && strings.Compare(previous, value.LogicalID) >= 0 {
			return fmt.Errorf("logical identifiers are not sorted and unique")
		}
		if err := validateIdentifierV2(value, kind, ""); err != nil {
			return err
		}
		if _, exists := seenNames[value.PhysicalName]; exists {
			return fmt.Errorf("physical identifier is duplicated")
		}
		seenNames[value.PhysicalName] = struct{}{}
		previous = value.LogicalID
	}
	return nil
}

func validateFieldsV2(
	fields []schemav2.FieldIdentifier,
	seenNames map[string]struct{},
) error {
	previousKey := ""
	present := make(map[fieldKey]bool, len(fields))
	for _, field := range fields {
		expected, err := schemav2.DeriveFieldIdentifier(
			field.LogicalID, field.Generation, field.Variant,
		)
		if err != nil || field.PhysicalName != expected {
			return fmt.Errorf("field physical identifier is not deterministic")
		}
		key := field.LogicalID + ":" + fmt.Sprintf("%010d", field.Generation) +
			":" + variantOrder(field.Variant)
		if previousKey != "" && previousKey >= key {
			return fmt.Errorf("field identifiers are not sorted and unique")
		}
		if _, exists := seenNames[field.PhysicalName]; exists {
			return fmt.Errorf("field physical identifier is duplicated")
		}
		seenNames[field.PhysicalName] = struct{}{}
		present[fieldKey{field.LogicalID, field.Generation, field.Variant}] = true
		previousKey = key
	}
	for key := range present {
		if key.variant == schemav2.VariantAccent &&
			!present[fieldKey{key.id, key.generation, schemav2.VariantFolded}] {
			return fmt.Errorf("accent field mapping requires folded companion")
		}
	}
	return nil
}

type fieldKey struct {
	id         string
	generation int64
	variant    schemav2.FieldVariant
}

func variantOrder(variant schemav2.FieldVariant) string {
	switch variant {
	case schemav2.VariantValue:
		return "0"
	case schemav2.VariantFolded:
		return "1"
	case schemav2.VariantAccent:
		return "2"
	default:
		return "9"
	}
}
