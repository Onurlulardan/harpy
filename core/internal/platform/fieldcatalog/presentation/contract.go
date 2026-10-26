package presentation

import "encoding/json"

const (
	ContractVersion         = "forms-presentation.v1"
	RegistryContractVersion = "forms-presentation-registry.v1"
	FieldCatalogContract    = "forms-presentation-field-catalog.v1"
	AttestationContract     = "forms-presentation-attestation.v1"
	ValidatorVersion        = "1.0.0"
)

type Document struct {
	RegistryContract string    `json:"registry_contract"`
	DefaultLocale    string    `json:"default_locale"`
	Locales          []string  `json:"locales"`
	Surfaces         []Surface `json:"surfaces"`
}

type Surface struct {
	ID           string        `json:"id"`
	Kind         string        `json:"kind"`
	Translations []Translation `json:"translations"`
	Root         Node          `json:"root"`
}

type Node struct {
	ID           string        `json:"id"`
	Kind         string        `json:"kind"`
	Variant      string        `json:"variant,omitempty"`
	FieldID      string        `json:"field_id,omitempty"`
	Widget       string        `json:"widget,omitempty"`
	Interaction  string        `json:"interaction,omitempty"`
	Span         int           `json:"span"`
	Columns      int           `json:"columns,omitempty"`
	Translations []Translation `json:"translations"`
	Visibility   *Visibility   `json:"visibility,omitempty"`
	Children     []Node        `json:"children,omitempty"`
}

type Translation struct {
	Locale      string `json:"locale"`
	Label       string `json:"label,omitempty"`
	Help        string `json:"help,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
}

type Visibility struct {
	Match      string      `json:"match"`
	Conditions []Condition `json:"conditions"`
}

type Condition struct {
	FieldID  string          `json:"field_id"`
	Operator string          `json:"operator"`
	Value    json.RawMessage `json:"value,omitempty"`
}
