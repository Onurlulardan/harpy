package schemav2

type VerifyArguments struct {
	Checks      []VerifyCheck              `json:"checks"`
	Fields      []VerifyFieldContract      `json:"fields"`
	Relations   []RelationArguments        `json:"relations"`
	Indexes     []VerifyIndexContract      `json:"indexes"`
	Constraints []VerifyConstraintContract `json:"constraints"`
	Evolution   []EvolutionDataContract    `json:"evolution"`
}

type VerifyFieldContract struct {
	FieldID          string       `json:"field_id"`
	Generation       int64        `json:"generation"`
	Variant          FieldVariant `json:"variant"`
	StorageType      StorageType  `json:"storage_type"`
	Nullable         bool         `json:"nullable"`
	DecimalPrecision *int         `json:"decimal_precision,omitempty"`
	DecimalScale     *int         `json:"decimal_scale,omitempty"`
}

type VerifyIndexContract struct {
	IndexID           string             `json:"index_id"`
	Generation        int64              `json:"generation"`
	Fields            []IndexField       `json:"fields"`
	Method            IndexMethod        `json:"method"`
	OperatorClass     IndexOperatorClass `json:"operator_class"`
	Unique            bool               `json:"unique"`
	ActiveRecordsOnly bool               `json:"active_records_only"`
}

type EvolutionDataContract struct {
	Source         FieldReference    `json:"source"`
	Target         FieldReference    `json:"target"`
	Converter      Converter         `json:"converter"`
	Checksum       ChecksumAlgorithm `json:"checksum"`
	DefaultLiteral *TypedLiteral     `json:"default_literal,omitempty"`
}

type VerifyConstraintContract struct {
	ConstraintID string           `json:"constraint_id"`
	Generation   int64            `json:"generation"`
	Kind         ConstraintKind   `json:"kind"`
	Fields       []FieldReference `json:"fields"`
	Minimum      *int             `json:"minimum,omitempty"`
	Maximum      *int             `json:"maximum,omitempty"`
}
