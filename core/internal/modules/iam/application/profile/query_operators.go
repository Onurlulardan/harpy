package profile

import restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func queryOperators(field Field) []restquery.Operator {
	null := []restquery.Operator{
		restquery.OperatorIsNull, restquery.OperatorIsNotNull,
	}
	if field.cardinality == CardinalityList {
		return append([]restquery.Operator{
			restquery.OperatorContainsAny, restquery.OperatorContainsAll,
		}, null...)
	}
	if field.kind == LogicalJSON {
		return append([]restquery.Operator{
			restquery.OperatorEq, restquery.OperatorNe,
		}, null...)
	}
	return append([]restquery.Operator{
		restquery.OperatorEq, restquery.OperatorNe,
		restquery.OperatorIn, restquery.OperatorNotIn,
	}, null...)
}
