package recordkernel

import "encoding/json"

const (
	StorageSchemaContractVersion                = "forms-storage-schema.v1"
	PresentationContractVersion                 = "forms-presentation.v1"
	FeatureBaseColumn            StorageFeature = "base_column"
)

type StorageSchema struct {
	FieldRegistryContract string                  `json:"field_registry_contract"`
	Fields                []FieldDefinition       `json:"fields"`
	Relations             []RelationDefinition    `json:"relations"`
	Indexes               []IndexDefinition       `json:"indexes"`
	Constraints           []ConstraintDefinition  `json:"constraints"`
	Validations           []ValidationDefinition  `json:"validations"`
	Calculations          []CalculationDefinition `json:"calculations"`
}

type FieldDefinition struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	TypeID          TypeID             `json:"type"`
	Role            Role               `json:"role"`
	Nullable        bool               `json:"nullable"`
	StorageFeatures []StorageFeature   `json:"storage_features"`
	Decimal         *DecimalDefinition `json:"decimal,omitempty"`
	EnumOptions     []string           `json:"enum_options,omitempty"`
}

type DecimalDefinition struct {
	Precision int `json:"precision"`
	Scale     int `json:"scale"`
}

type RelationDefinition struct {
	ID                      string  `json:"id"`
	FieldID                 string  `json:"field_id"`
	TargetFormID            string  `json:"target_form_id"`
	SourceConstraintID      string  `json:"source_constraint_id"`
	TargetConstraintID      string  `json:"target_constraint_id"`
	CardinalityConstraintID *string `json:"cardinality_constraint_id,omitempty"`
	Cardinality             string  `json:"cardinality"`
	DeletePolicy            string  `json:"delete_policy"`
}

type IndexDefinition struct {
	ID                string   `json:"id"`
	FieldIDs          []string `json:"field_ids"`
	Method            string   `json:"method"`
	Unique            bool     `json:"unique"`
	ActiveRecordsOnly bool     `json:"active_records_only"`
}

type ConstraintDefinition struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	FieldIDs []string `json:"field_ids"`
	Minimum  *int     `json:"minimum,omitempty"`
	Maximum  *int     `json:"maximum,omitempty"`
}

type ValidationDefinition struct {
	Kind     string   `json:"kind"`
	FieldIDs []string `json:"field_ids"`
	Minimum  *int     `json:"minimum,omitempty"`
	Maximum  *int     `json:"maximum,omitempty"`
}

type CalculationDefinition struct {
	FieldID    string     `json:"field_id"`
	Expression Expression `json:"expression"`
}

type Expression struct {
	Kind      string          `json:"kind"`
	FieldID   *string         `json:"field_id,omitempty"`
	Literal   json.RawMessage `json:"literal,omitempty"`
	Operator  *string         `json:"operator,omitempty"`
	Function  *string         `json:"function,omitempty"`
	Arguments []Expression    `json:"arguments,omitempty"`
}
