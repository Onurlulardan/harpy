package nativequery

import "encoding/json"

const ContractV1 = "analytics.native-join-query.v1"

type Plan struct {
	Contract    string       `json:"contract"`
	Sources     []Source     `json:"sources"`
	Joins       []Join       `json:"joins"`
	Fields      []Field      `json:"fields"`
	Aggregation *Aggregation `json:"aggregation,omitempty"`
	Order       []Order      `json:"order,omitempty"`
	Limit       int          `json:"limit"`
}

type Source struct {
	Alias    string          `json:"alias"`
	FormSlug string          `json:"form_slug"`
	Query    json.RawMessage `json:"query"`
}

type Join struct {
	Kind        string `json:"kind"`
	Source      string `json:"source"`
	SourceField string `json:"source_field"`
	Target      string `json:"target"`
	TargetField string `json:"target_field"`
}

type Field struct {
	Alias  string `json:"alias"`
	Source string `json:"source"`
	Field  string `json:"field"`
}

type Aggregation struct {
	Groups   []Field   `json:"groups"`
	Measures []Measure `json:"measures"`
}

type Measure struct {
	Alias    string `json:"alias"`
	Source   string `json:"source,omitempty"`
	Field    string `json:"field,omitempty"`
	Function string `json:"function"`
}

type Order struct {
	Source    string `json:"source,omitempty"`
	Field     string `json:"field,omitempty"`
	Measure   string `json:"measure,omitempty"`
	Direction string `json:"direction"`
}
