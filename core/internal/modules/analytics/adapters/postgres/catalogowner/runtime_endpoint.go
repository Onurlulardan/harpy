package catalogowner

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/httpdecode"
)

type endpointDefinition struct {
	DataSources []endpointSource `json:"dataSources"`
	Fields      []endpointField  `json:"fields"`
	Filters     []endpointFilter `json:"filters"`
	Joins       []endpointJoin   `json:"joins"`
	Aggregation struct {
		Enabled    bool                `json:"enabled"`
		Aggregates []endpointAggregate `json:"aggregates"`
		GroupBy    []string            `json:"groupBy"`
		OrderBy    []endpointOrder     `json:"orderBy"`
	} `json:"aggregation"`
	Response struct {
		Format  string          `json:"format"`
		OrderBy []endpointOrder `json:"orderBy"`
	} `json:"response"`
}

type endpointSource struct {
	Alias    string `json:"alias"`
	FormSlug string `json:"formSlug"`
}

type endpointField struct {
	Alias  string `json:"alias"`
	Field  string `json:"field"`
	Source string `json:"source"`
}

type endpointFilter struct {
	Field    string            `json:"field"`
	Operator string            `json:"operator"`
	Value    json.RawMessage   `json:"value"`
	Values   []json.RawMessage `json:"values"`
}

type endpointAggregate struct {
	Alias    string `json:"alias"`
	Field    string `json:"field"`
	Function string `json:"function"`
}

type endpointOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type endpointJoin struct {
	JoinType    string `json:"joinType"`
	SourceField string `json:"sourceField"`
	SourceForm  string `json:"sourceForm"`
	TargetField string `json:"targetField"`
	TargetForm  string `json:"targetForm"`
}

func endpointRuntimePlan(document []byte) (runtimePlan, error) {
	plan := runtimePlan{Contract: runtimePlanContract, Kind: "endpoint",
		Parameters: json.RawMessage(`{}`), Output: json.RawMessage(`{}`)}
	var definition endpointDefinition
	if err := json.Unmarshal(document, &definition); err != nil ||
		len(definition.DataSources) == 0 {
		return runtimePlan{}, fmt.Errorf("Analytics endpoint definition is invalid")
	}
	for _, source := range definition.DataSources {
		if source.FormSlug == "" || source.Alias == "" {
			return runtimePlan{}, fmt.Errorf("Analytics endpoint source is invalid")
		}
		plan.Sources = append(plan.Sources, source.FormSlug)
	}
	if len(definition.DataSources) != 1 {
		query, err := nativeJoinRuntimeQuery(definition)
		if err != nil {
			return runtimePlan{}, err
		}
		plan.Query = query
		plan.Status = "candidate"
		return plan, nil
	}
	query, err := endpointQuery(definition, definition.DataSources[0].Alias)
	if err != nil {
		plan.Status = "pending"
		plan.BlockerCode = "ANALYTICS_QUERY_CONTRACT_UNSUPPORTED"
		return plan, nil
	}
	if _, err = httpdecode.POST().Decode(query); err != nil {
		plan.Status = "pending"
		plan.BlockerCode = "ANALYTICS_QUERY_CONTRACT_INVALID"
		return plan, nil
	}
	plan.Status = "candidate"
	plan.Query = query
	return plan, nil
}

func endpointQuery(definition endpointDefinition, alias string) ([]byte, error) {
	where, err := endpointWhere(definition.Filters, alias)
	if err != nil {
		return nil, err
	}
	root := map[string]any{}
	if where != nil {
		root["where"] = where
	}
	if definition.Aggregation.Enabled {
		aggregate, buildErr := endpointAggregation(definition, alias)
		if buildErr != nil {
			return nil, buildErr
		}
		root["aggregate"] = aggregate
	} else {
		fields := make([]string, 0, len(definition.Fields))
		for _, field := range definition.Fields {
			if field.Source != alias {
				return nil, fmt.Errorf("endpoint field requires a native join")
			}
			name, ok := unqualified(field.Field, alias)
			if !ok {
				return nil, fmt.Errorf("endpoint field is invalid")
			}
			fields = append(fields, name)
		}
		root["select"] = fields
		root["page"] = map[string]any{"size": 50, "after": nil}
		root["count"] = "none"
		root["include"] = []any{}
	}
	return json.Marshal(root)
}

func endpointWhere(filters []endpointFilter, alias string) (any, error) {
	predicates := make([]any, 0, len(filters))
	for _, filter := range filters {
		field, ok := unqualified(filter.Field, alias)
		operator, opOK := endpointOperator(filter.Operator)
		if !ok || !opOK {
			return nil, fmt.Errorf("endpoint filter is unsupported")
		}
		predicate := map[string]any{"field": field, "op": operator}
		if operator == "in" || operator == "not_in" {
			if len(filter.Values) == 0 {
				return nil, fmt.Errorf("endpoint filter values are empty")
			}
			predicate["values"] = filter.Values
		} else {
			if len(filter.Value) == 0 {
				return nil, fmt.Errorf("endpoint filter value is empty")
			}
			predicate["value"] = filter.Value
		}
		predicates = append(predicates, predicate)
	}
	if len(predicates) == 0 {
		return nil, nil
	}
	if len(predicates) == 1 {
		return predicates[0], nil
	}
	return map[string]any{"and": predicates}, nil
}

func endpointOperator(value string) (string, bool) {
	switch strings.ToUpper(value) {
	case "=":
		return "eq", true
	case "!=":
		return "ne", true
	case "IN":
		return "in", true
	case "NOT_IN":
		return "not_in", true
	default:
		return "", false
	}
}
