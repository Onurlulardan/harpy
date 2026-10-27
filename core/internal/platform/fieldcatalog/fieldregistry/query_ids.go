package fieldregistry

const (
	QueryText      QueryType = "text"
	QueryInteger   QueryType = "integer"
	QueryDecimal   QueryType = "decimal"
	QueryBoolean   QueryType = "boolean"
	QueryDate      QueryType = "date"
	QueryTimestamp QueryType = "timestamp"
	QueryUUID      QueryType = "uuid"
	QueryEnum      QueryType = "enum"
	QueryRelation  QueryType = "relation"
	QueryTextList  QueryType = "text_list"
	QueryUUIDList  QueryType = "uuid_list"
	QueryJSON      QueryType = "json"
)

const (
	GateUnavailable       Gate = "unavailable"
	GateAlways            Gate = "always"
	GateTextCasefold      Gate = "text_casefold_v1"
	GateEnumOrdinal       Gate = "enum_ordinal"
	GateRelationToOne     Gate = "relation_to_one"
	GateJSONEqualityIndex Gate = "json_equality_index"
)

const (
	OperatorEq          Operator = "eq"
	OperatorNe          Operator = "ne"
	OperatorGT          Operator = "gt"
	OperatorGTE         Operator = "gte"
	OperatorLT          Operator = "lt"
	OperatorLTE         Operator = "lte"
	OperatorContains    Operator = "contains"
	OperatorStartsWith  Operator = "starts_with"
	OperatorEndsWith    Operator = "ends_with"
	OperatorIn          Operator = "in"
	OperatorNotIn       Operator = "not_in"
	OperatorBetween     Operator = "between"
	OperatorIsNull      Operator = "is_null"
	OperatorIsNotNull   Operator = "is_not_null"
	OperatorContainsAny Operator = "contains_any"
	OperatorContainsAll Operator = "contains_all"
)

const (
	AggregateCountDistinct AggregateFunction = "count_distinct"
	AggregateSum           AggregateFunction = "sum"
	AggregateAvg           AggregateFunction = "avg"
	AggregateMin           AggregateFunction = "min"
	AggregateMax           AggregateFunction = "max"
)
