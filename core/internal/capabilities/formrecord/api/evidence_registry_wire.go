package recordkernel

type evidenceStorage struct {
	Type            string   `json:"type"`
	Layout          string   `json:"layout"`
	AllowedFeatures []string `json:"allowed_features"`
}

type evidenceOperator struct {
	Operator  string `json:"operator"`
	Gate      string `json:"gate"`
	CountSafe bool   `json:"count_safe"`
}

type evidenceAggregate struct {
	Function string `json:"function"`
	Gate     string `json:"gate"`
}

type evidenceQuery struct {
	LogicalType         string              `json:"logical_type"`
	SelectGate          string              `json:"select_gate"`
	SortGate            string              `json:"sort_gate"`
	SearchGate          string              `json:"search_gate"`
	GroupGate           string              `json:"group_gate"`
	RequiresCursorBound bool                `json:"requires_cursor_bound"`
	Operators           []evidenceOperator  `json:"operators"`
	Aggregates          []evidenceAggregate `json:"aggregates"`
}

type evidenceCost struct {
	Class      string   `json:"class"`
	Dimensions []string `json:"dimensions"`
}

type evidenceDescriptor struct {
	TypeID       string          `json:"type_id"`
	ValueProfile string          `json:"value_profile"`
	Roles        []string        `json:"roles"`
	Storage      evidenceStorage `json:"storage"`
	Query        evidenceQuery   `json:"query"`
	Cost         evidenceCost    `json:"cost"`
}

func descriptorEvidence(value Descriptor) evidenceDescriptor {
	roles := make([]string, len(value.Roles))
	for index, role := range value.Roles {
		roles[index] = string(role)
	}
	features := make([]string, len(value.Storage.AllowedFeatures))
	for index, feature := range value.Storage.AllowedFeatures {
		features[index] = string(feature)
	}
	operators := make([]evidenceOperator, len(value.Query.Operators))
	for index, item := range value.Query.Operators {
		operators[index] = evidenceOperator{string(item.Operator), string(item.Gate), item.CountSafe}
	}
	aggregates := make([]evidenceAggregate, len(value.Query.Aggregates))
	for index, item := range value.Query.Aggregates {
		aggregates[index] = evidenceAggregate{string(item.Function), string(item.Gate)}
	}
	dimensions := make([]string, len(value.Cost.Dimensions))
	for index, item := range value.Cost.Dimensions {
		dimensions[index] = string(item)
	}
	return evidenceDescriptor{string(value.TypeID), string(value.ValueProfile), roles,
		evidenceStorage{string(value.Storage.Type), string(value.Storage.Layout), features},
		evidenceQuery{string(value.Query.LogicalType), string(value.Query.SelectGate),
			string(value.Query.SortGate), string(value.Query.SearchGate),
			string(value.Query.GroupGate), value.Query.RequiresCursorBound, operators, aggregates},
		evidenceCost{string(value.Cost.Class), dimensions}}
}
