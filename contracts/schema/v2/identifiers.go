package schemav2

type Identifier struct {
	LogicalID    string `json:"logical_id"`
	Generation   int64  `json:"generation,omitempty"`
	PhysicalName string `json:"physical_name"`
}

type FieldIdentifier struct {
	LogicalID    string       `json:"logical_id"`
	Generation   int64        `json:"generation"`
	Variant      FieldVariant `json:"variant"`
	PhysicalName string       `json:"physical_name"`
}

type IdentifierMapping struct {
	Form         Identifier        `json:"form"`
	RelatedForms []Identifier      `json:"related_forms"`
	Fields       []FieldIdentifier `json:"fields"`
	Relations    []Identifier      `json:"relations"`
	Indexes      []Identifier      `json:"indexes"`
	Constraints  []Identifier      `json:"constraints"`
}
