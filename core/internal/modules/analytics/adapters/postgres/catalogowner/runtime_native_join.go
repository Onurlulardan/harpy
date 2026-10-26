package catalogowner

import (
	"fmt"
	"strings"

	nativequery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/nativequery"
)

func nativeJoinRuntimeQuery(definition endpointDefinition) ([]byte, error) {
	fields, err := nativeJoinFields(definition)
	if err != nil {
		return nil, err
	}
	sources, err := nativeJoinSources(definition, fields)
	if err != nil {
		return nil, err
	}
	joins := make([]nativequery.Join, len(definition.Joins))
	if len(joins) != len(sources)-1 {
		return nil, fmt.Errorf("Analytics native join graph is incomplete")
	}
	for index, value := range definition.Joins {
		sourceField, sourceOK := unqualified(value.SourceField, value.SourceForm)
		targetField, targetOK := unqualified(value.TargetField, value.TargetForm)
		if !sourceOK || !targetOK || strings.ToLower(value.JoinType) != "left" {
			return nil, fmt.Errorf("Analytics native join is unsupported")
		}
		joins[index] = nativequery.Join{Kind: "left", Source: value.SourceForm,
			SourceField: sourceField, Target: value.TargetForm, TargetField: targetField}
	}
	aggregation, err := nativeJoinAggregation(definition, fields)
	if err != nil {
		return nil, err
	}
	orders := definition.Response.OrderBy
	if aggregation != nil {
		orders = definition.Aggregation.OrderBy
	}
	normalizedOrders, err := nativeJoinOrders(orders, fields, aggregation)
	if err != nil {
		return nil, err
	}
	limit := 50
	if aggregation != nil {
		limit = 100
	}
	return nativequery.Encode(nativequery.Plan{
		Contract: nativequery.ContractV1, Sources: sources, Joins: joins,
		Fields: fields, Aggregation: aggregation, Order: normalizedOrders, Limit: limit,
	})
}

func nativeJoinFields(
	definition endpointDefinition,
) ([]nativequery.Field, error) {
	result := make([]nativequery.Field, len(definition.Fields))
	aliases := map[string]bool{}
	for index, value := range definition.Fields {
		field, ok := unqualified(value.Field, value.Source)
		if !ok || value.Alias == "" || aliases[value.Alias] {
			return nil, fmt.Errorf("Analytics native join field is invalid")
		}
		aliases[value.Alias] = true
		result[index] = nativequery.Field{
			Alias: value.Alias, Source: value.Source, Field: field,
		}
	}
	return result, nil
}
