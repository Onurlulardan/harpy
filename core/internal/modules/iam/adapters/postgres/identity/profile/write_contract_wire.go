package profilepostgres

import "encoding/json"

type profileWriteContractWire struct {
	Contract string                    `json:"contract"`
	PlanID   *string                   `json:"plan_id"`
	Source   json.RawMessage           `json:"source"`
	Target   json.RawMessage           `json:"target"`
	Mappings []profileWriteMappingWire `json:"mappings"`
}

type profileWriteMappingWire struct {
	SourceFieldID string `json:"source_field_id"`
	TargetFieldID string `json:"target_field_id"`
	Converter     string `json:"converter_kind"`
	SemanticSHA   string `json:"semantic_sha256"`
}
