package authorization

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func mandatoryArityValid(operator api.Operator, count int) bool {
	arity, valid := operator.Arity()
	if !valid {
		return false
	}
	switch arity {
	case api.PredicateArityZero:
		return count == 0
	case api.PredicateArityOne:
		return count == 1
	case api.PredicateArityList:
		return count >= 1 && count <= maximumMandatoryLiterals
	case api.PredicateArityPair:
		return count == 2
	default:
		return false
	}
}

func operatorPossibleForLogicalType(
	logicalType api.LogicalType,
	operator api.Operator,
) bool {
	if operator == api.OperatorIsNull || operator == api.OperatorIsNotNull {
		return true
	}
	switch logicalType {
	case api.LogicalTypeText:
		return scalarOrderOperator(operator) || patternOperator(operator)
	case api.LogicalTypeInteger, api.LogicalTypeDecimal,
		api.LogicalTypeDate, api.LogicalTypeTimestamp, api.LogicalTypeEnum:
		return scalarOrderOperator(operator)
	case api.LogicalTypeBoolean, api.LogicalTypeUUID, api.LogicalTypeRelation:
		return equalityOperator(operator)
	case api.LogicalTypeTextList, api.LogicalTypeUUIDList:
		return operator == api.OperatorContainsAny || operator == api.OperatorContainsAll
	case api.LogicalTypeJSON:
		return operator == api.OperatorEq || operator == api.OperatorNe
	default:
		return false
	}
}

func scalarOrderOperator(operator api.Operator) bool {
	return equalityOperator(operator) || operator == api.OperatorGT ||
		operator == api.OperatorGTE || operator == api.OperatorLT ||
		operator == api.OperatorLTE || operator == api.OperatorBetween
}

func equalityOperator(operator api.Operator) bool {
	return operator == api.OperatorEq || operator == api.OperatorNe ||
		operator == api.OperatorIn || operator == api.OperatorNotIn
}

func patternOperator(operator api.Operator) bool {
	return operator == api.OperatorContains || operator == api.OperatorStartsWith ||
		operator == api.OperatorEndsWith
}

func mandatoryValueKind(logicalType api.LogicalType) api.LogicalType {
	if logicalType == api.LogicalTypeTextList {
		return api.LogicalTypeText
	}
	if logicalType == api.LogicalTypeUUIDList {
		return api.LogicalTypeUUID
	}
	return logicalType
}
