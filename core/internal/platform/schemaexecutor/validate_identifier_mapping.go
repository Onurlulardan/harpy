package schemaexecutor

import (
	"fmt"
	"strings"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
)

func validateIdentifierMapping(
	mapping schemav1.IdentifierMapping,
	resourceID string,
) error {
	if mapping.RelatedForms == nil || mapping.Fields == nil ||
		mapping.Relations == nil || mapping.Indexes == nil || mapping.Constraints == nil {
		return fmt.Errorf("identifier mapping arrays must be explicit")
	}
	if err := validateIdentifier(mapping.Form, "frm_", resourceID); err != nil {
		return fmt.Errorf("form mapping: %w", err)
	}
	if len(mapping.Fields) > 1200 || len(mapping.Relations) > 100 ||
		len(mapping.Indexes) > 128 || len(mapping.Constraints) > 256 {
		return fmt.Errorf("identifier mapping exceeds object budget")
	}
	seenNames := map[string]struct{}{mapping.Form.PhysicalName: {}}
	if err := validateIdentifiers(mapping.RelatedForms, "frm_", seenNames); err != nil {
		return fmt.Errorf("related form mapping: %w", err)
	}
	if err := validateFields(mapping.Fields, seenNames); err != nil {
		return err
	}
	for label, values := range map[string]struct {
		prefix string
		items  []schemav1.Identifier
	}{
		"relation":   {"rel_", mapping.Relations},
		"index":      {"idx_", mapping.Indexes},
		"constraint": {"cst_", mapping.Constraints},
	} {
		if err := validateIdentifiers(values.items, values.prefix, seenNames); err != nil {
			return fmt.Errorf("%s mapping: %w", label, err)
		}
	}
	return nil
}

func validateIdentifier(value schemav1.Identifier, prefix, expectedID string) error {
	if expectedID != "" && value.LogicalID != expectedID {
		return fmt.Errorf("logical ID differs from resource")
	}
	expected, err := deriveObjectIdentifier(prefix, value.LogicalID)
	if err != nil || value.PhysicalName != expected {
		return fmt.Errorf("physical identifier is not deterministic")
	}
	return nil
}

func validateIdentifiers(
	values []schemav1.Identifier,
	prefix string,
	seenNames map[string]struct{},
) error {
	previous := ""
	for _, value := range values {
		if previous != "" && strings.Compare(previous, value.LogicalID) >= 0 {
			return fmt.Errorf("logical identifiers are not sorted and unique")
		}
		if err := validateIdentifier(value, prefix, ""); err != nil {
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
