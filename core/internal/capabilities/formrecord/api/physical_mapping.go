package recordkernel

import (
	"strings"

	"github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/google/uuid"
)

type PhysicalMapping struct {
	LogicalID    uuid.UUID `json:"logical_id"`
	Kind         string    `json:"kind"`
	PhysicalName string    `json:"physical_name"`
	StorageType  string    `json:"storage_type"`
	Variant      string    `json:"variant"`
	Generation   uint32    `json:"generation"`
	Profile      string    `json:"profile"`
}

func (mapping PhysicalMapping) Valid() bool {
	return stableUUIDv7(mapping.LogicalID) && mapping.Generation > 0 &&
		validPhysicalShape(mapping) && exactPhysicalName(mapping)
}

func validPhysicalShape(mapping PhysicalMapping) bool {
	if mapping.Kind != "field_column" {
		wantStorage := map[string]string{
			"form_table": "record", "relation_table": "relation",
			"index": "index", "constraint": "constraint",
		}[mapping.Kind]
		return mapping.Variant == "object" && wantStorage != "" &&
			mapping.StorageType == wantStorage && validObjectProfile(mapping.Kind, mapping.Profile)
	}
	if !validFieldStorageType(mapping.StorageType) {
		return false
	}
	switch mapping.Variant {
	case "value":
		return !strings.HasSuffix(mapping.PhysicalName, "_folded") &&
			!strings.HasSuffix(mapping.PhysicalName, "_accent") &&
			validValueProfile(mapping.StorageType, mapping.Profile)
	case "folded", "accent":
		profile := map[string]string{"folded": "unicode_default_casefold_v1",
			"accent": "accent_insensitive_v1"}[mapping.Variant]
		return mapping.StorageType == "text" && mapping.Profile == profile &&
			strings.HasSuffix(mapping.PhysicalName, "_"+mapping.Variant)
	default:
		return false
	}
}

func exactPhysicalName(mapping PhysicalMapping) bool {
	logicalID := mapping.LogicalID.String()
	var expected string
	var err error
	switch mapping.Kind {
	case "form_table":
		expected, err = schemav2.DeriveIdentifier(schemav2.IdentifierForm, logicalID)
	case "relation_table":
		expected, err = schemav2.DeriveIdentifier(schemav2.IdentifierRelation, logicalID)
	case "field_column":
		expected, err = schemav2.DeriveFieldIdentifier(
			logicalID, int64(mapping.Generation), schemav2.FieldVariant(mapping.Variant),
		)
	case "index", "constraint":
		return validAuxiliaryName(mapping)
	default:
		return false
	}
	return err == nil && expected == mapping.PhysicalName
}

func validAuxiliaryName(mapping PhysicalMapping) bool {
	if mapping.Kind == "index" {
		expected, err := schemav2.DeriveIndexIdentifier(
			mapping.LogicalID.String(), int64(mapping.Generation),
		)
		return err == nil && expected == mapping.PhysicalName
	}
	expected, err := schemav2.DeriveIdentifier(
		schemav2.IdentifierConstraint, mapping.LogicalID.String(),
	)
	return err == nil && expected == mapping.PhysicalName
}

func validFieldStorageType(value string) bool {
	switch value {
	case "text", "integer", "decimal", "boolean", "date", "timestamp",
		"uuid", "enum", "text_list", "uuid_list", "json":
		return true
	default:
		return false
	}
}

func validValueProfile(storageType, profile string) bool {
	allowed := map[string]map[string]bool{
		"text":    {"text_nfc_v1": true, "time_microsecond_v1": true},
		"integer": {"integer_base10_v1": true}, "decimal": {"decimal_base10_v1": true},
		"boolean": {"boolean_v1": true}, "date": {"date_iso8601_v1": true},
		"timestamp": {"timestamp_utc_microsecond_v1": true},
		"uuid":      {"uuid_rfc4122_lower_v1": true, "document_ref_uuid_v1": true},
		"enum":      {"enum_code_v1": true},
		"text_list": {"text_list_nfc_v1": true, "enum_list_codes_v1": true},
		"uuid_list": {"uuid_list_v1": true, "document_ref_list_v1": true},
		"json":      {"json_rfc8785_v1": true},
	}
	return allowed[storageType][profile]
}

func validObjectProfile(kind, profile string) bool {
	allowed := map[string]map[string]bool{
		"form_table":     {"forms_record_table_v2": true},
		"relation_table": {"forms_relation_edge_v1": true},
		"index": {"forms_index_btree_default_v1": true,
			"forms_index_gin_default_v1": true, "forms_index_gin_trgm_v1": true},
		"constraint": {"forms_constraint_not_null_v1": true,
			"forms_constraint_unique_v1": true, "forms_constraint_text_length_v1": true,
			"forms_relation_source_constraint_v1":      true,
			"forms_relation_target_constraint_v1":      true,
			"forms_relation_cardinality_constraint_v1": true},
	}
	return allowed[kind][profile]
}
