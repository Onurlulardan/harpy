package api

type Translation struct {
	Help   string `json:"help"`
	Label  string `json:"label"`
	Locale string `json:"locale"`
}

type Field struct {
	FieldID      string        `json:"field_id"`
	Interaction  Interaction   `json:"interaction"`
	Order        int32         `json:"order"`
	Translations []Translation `json:"translations"`
	Widget       Widget        `json:"widget"`
}

type Section struct {
	Fields       []Field       `json:"fields"`
	ID           string        `json:"id"`
	Order        int32         `json:"order"`
	Translations []Translation `json:"translations"`
}

type DefinitionContract string

const (
	DefinitionContractV1 DefinitionContract = "harpy/system-resource-layout/v1"
	DefinitionContractV2 DefinitionContract = "harpy/system-resource-layout/v2"
)

type Definition struct {
	SourceContract DefinitionContract  `json:"-"`
	Surfaces       []SurfaceDefinition `json:"surfaces"`
}
