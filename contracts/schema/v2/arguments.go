package schemav2

import "encoding/json"

type OperationArguments map[string]json.RawMessage

type EmptyArguments struct{}

type ColumnArguments struct {
	FieldID          string       `json:"field_id"`
	Generation       int64        `json:"generation"`
	Variant          FieldVariant `json:"variant"`
	StorageType      StorageType  `json:"storage_type"`
	Nullable         bool         `json:"nullable"`
	DecimalPrecision *int         `json:"decimal_precision,omitempty"`
	DecimalScale     *int         `json:"decimal_scale,omitempty"`
}

type FieldArguments struct {
	FieldID    string       `json:"field_id"`
	Generation int64        `json:"generation"`
	Variant    FieldVariant `json:"variant"`
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

type FieldReference struct {
	FieldID    string       `json:"field_id"`
	Generation int64        `json:"generation"`
	Variant    FieldVariant `json:"variant"`
}

type IndexField struct {
	FieldID     string       `json:"field_id"`
	Generation  int64        `json:"generation"`
	Variant     FieldVariant `json:"variant"`
	StorageType StorageType  `json:"storage_type"`
}

type IndexArguments struct {
	IndexID           string             `json:"index_id"`
	Generation        int64              `json:"generation"`
	Fields            []IndexField       `json:"fields"`
	Method            IndexMethod        `json:"method"`
	OperatorClass     IndexOperatorClass `json:"operator_class"`
	Unique            bool               `json:"unique"`
	ActiveRecordsOnly bool               `json:"active_records_only"`
}

type ConstraintArguments struct {
	ConstraintID string           `json:"constraint_id"`
	Generation   int64            `json:"generation"`
	Kind         ConstraintKind   `json:"kind"`
	Fields       []FieldReference `json:"fields"`
	Minimum      *int             `json:"minimum,omitempty"`
	Maximum      *int             `json:"maximum,omitempty"`
}

type ObjectArguments struct {
	ObjectID   string `json:"object_id"`
	Generation *int64 `json:"generation,omitempty"`
}

type BackfillArguments struct {
	Source         FieldReference `json:"source"`
	Target         FieldReference `json:"target"`
	Converter      Converter      `json:"converter"`
	DefaultLiteral *TypedLiteral  `json:"default_literal,omitempty"`
}

type TypedLiteral struct {
	StorageType StorageType     `json:"storage_type"`
	Value       json.RawMessage `json:"value"`
}

type GenerationArguments struct {
	FormVersionID    string `json:"form_version_id"`
	SourceGeneration int64  `json:"source_generation"`
	TargetGeneration int64  `json:"target_generation"`
}

type RetentionArguments struct {
	FormVersionID        string `json:"form_version_id"`
	RetainedGeneration   int64  `json:"retained_generation"`
	MinimumRetentionDays int    `json:"minimum_retention_days"`
	ZeroReferenceDays    int    `json:"zero_reference_days"`
}
