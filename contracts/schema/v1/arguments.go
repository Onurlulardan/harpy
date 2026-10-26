package schemav1

import "encoding/json"

type OperationArguments map[string]json.RawMessage

type EmptyArguments struct{}

type ColumnArguments struct {
	FieldID          string      `json:"field_id"`
	StorageType      StorageType `json:"storage_type"`
	Nullable         bool        `json:"nullable"`
	DecimalPrecision *int        `json:"decimal_precision,omitempty"`
	DecimalScale     *int        `json:"decimal_scale,omitempty"`
}

type FieldArguments struct {
	FieldID string `json:"field_id"`
}

type RelationArguments struct {
	RelationID              string       `json:"relation_id"`
	TargetFormID            string       `json:"target_form_id"`
	SourceConstraintID      string       `json:"source_constraint_id"`
	TargetConstraintID      string       `json:"target_constraint_id"`
	CardinalityConstraintID *string      `json:"cardinality_constraint_id,omitempty"`
	Cardinality             Cardinality  `json:"cardinality"`
	DeletePolicy            DeletePolicy `json:"delete_policy"`
}

type IndexArguments struct {
	IndexID           string      `json:"index_id"`
	FieldIDs          []string    `json:"field_ids"`
	Method            IndexMethod `json:"method"`
	Unique            bool        `json:"unique"`
	ActiveRecordsOnly bool        `json:"active_records_only"`
}

type ConstraintArguments struct {
	ConstraintID string         `json:"constraint_id"`
	Kind         ConstraintKind `json:"kind"`
	FieldIDs     []string       `json:"field_ids"`
	Minimum      *int           `json:"minimum,omitempty"`
	Maximum      *int           `json:"maximum,omitempty"`
}

type ObjectArguments struct {
	ObjectID string `json:"object_id"`
}

type BackfillArguments struct {
	SourceFieldID    string    `json:"source_field_id"`
	SourceGeneration int64     `json:"source_generation"`
	TargetFieldID    string    `json:"target_field_id"`
	TargetGeneration int64     `json:"target_generation"`
	Converter        Converter `json:"converter"`
}

type VerifyArguments struct {
	Checks      []VerifyCheck              `json:"checks"`
	Fields      []VerifyFieldContract      `json:"fields"`
	Relations   []RelationArguments        `json:"relations"`
	Indexes     []VerifyIndexContract      `json:"indexes"`
	Constraints []VerifyConstraintContract `json:"constraints"`
}

type FieldReference struct {
	FieldID    string `json:"field_id"`
	Generation int64  `json:"generation"`
}

type VerifyFieldContract struct {
	FieldID          string      `json:"field_id"`
	Generation       int64       `json:"generation"`
	StorageType      StorageType `json:"storage_type"`
	Nullable         bool        `json:"nullable"`
	DecimalPrecision *int        `json:"decimal_precision,omitempty"`
	DecimalScale     *int        `json:"decimal_scale,omitempty"`
}

type VerifyIndexContract struct {
	IndexID           string           `json:"index_id"`
	Fields            []FieldReference `json:"fields"`
	Method            IndexMethod      `json:"method"`
	Unique            bool             `json:"unique"`
	ActiveRecordsOnly bool             `json:"active_records_only"`
}

type VerifyConstraintContract struct {
	ConstraintID string           `json:"constraint_id"`
	Kind         ConstraintKind   `json:"kind"`
	Fields       []FieldReference `json:"fields"`
	Minimum      *int             `json:"minimum,omitempty"`
	Maximum      *int             `json:"maximum,omitempty"`
}

type GenerationArguments struct {
	FormVersionID    string `json:"form_version_id"`
	SourceGeneration int64  `json:"source_generation"`
	TargetGeneration int64  `json:"target_generation"`
}

type RetentionArguments struct {
	FormVersionID      string `json:"form_version_id"`
	RetainedGeneration int64  `json:"retained_generation"`
	ContractNotBefore  string `json:"contract_not_before"`
}
