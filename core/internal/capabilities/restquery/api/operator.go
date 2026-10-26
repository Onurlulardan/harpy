package api

type Operator uint8

const (
	OperatorEq Operator = iota + 1
	OperatorNe
	OperatorGT
	OperatorGTE
	OperatorLT
	OperatorLTE
	OperatorContains
	OperatorStartsWith
	OperatorEndsWith
	OperatorIn
	OperatorNotIn
	OperatorBetween
	OperatorIsNull
	OperatorIsNotNull
	OperatorContainsAny
	OperatorContainsAll
)

type PredicateArity uint8

const (
	PredicateArityZero PredicateArity = iota + 1
	PredicateArityOne
	PredicateArityList
	PredicateArityPair
)

func (Operator) MarshalJSON() ([]byte, error)       { return rejectDirectJSONMarshal() }
func (PredicateArity) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func ParseOperator(value string) (Operator, bool) {
	switch value {
	case "eq":
		return OperatorEq, true
	case "ne":
		return OperatorNe, true
	case "gt":
		return OperatorGT, true
	case "gte":
		return OperatorGTE, true
	case "lt":
		return OperatorLT, true
	case "lte":
		return OperatorLTE, true
	case "contains":
		return OperatorContains, true
	case "starts_with":
		return OperatorStartsWith, true
	case "ends_with":
		return OperatorEndsWith, true
	case "in":
		return OperatorIn, true
	case "not_in":
		return OperatorNotIn, true
	case "between":
		return OperatorBetween, true
	case "is_null":
		return OperatorIsNull, true
	case "is_not_null":
		return OperatorIsNotNull, true
	case "contains_any":
		return OperatorContainsAny, true
	case "contains_all":
		return OperatorContainsAll, true
	default:
		return 0, false
	}
}

func (operator Operator) String() string {
	for _, value := range canonicalOperators() {
		if value.operator == operator {
			return value.name
		}
	}
	return ""
}

func (operator Operator) Arity() (PredicateArity, bool) {
	switch operator {
	case OperatorIsNull, OperatorIsNotNull:
		return PredicateArityZero, true
	case OperatorEq, OperatorNe, OperatorGT, OperatorGTE, OperatorLT, OperatorLTE,
		OperatorContains, OperatorStartsWith, OperatorEndsWith:
		return PredicateArityOne, true
	case OperatorIn, OperatorNotIn, OperatorContainsAny, OperatorContainsAll:
		return PredicateArityList, true
	case OperatorBetween:
		return PredicateArityPair, true
	default:
		return 0, false
	}
}

type operatorName struct {
	operator Operator
	name     string
}

func canonicalOperators() []operatorName {
	return []operatorName{
		{OperatorEq, "eq"}, {OperatorNe, "ne"}, {OperatorGT, "gt"},
		{OperatorGTE, "gte"}, {OperatorLT, "lt"}, {OperatorLTE, "lte"},
		{OperatorContains, "contains"}, {OperatorStartsWith, "starts_with"},
		{OperatorEndsWith, "ends_with"}, {OperatorIn, "in"},
		{OperatorNotIn, "not_in"}, {OperatorBetween, "between"},
		{OperatorIsNull, "is_null"}, {OperatorIsNotNull, "is_not_null"},
		{OperatorContainsAny, "contains_any"}, {OperatorContainsAll, "contains_all"},
	}
}
