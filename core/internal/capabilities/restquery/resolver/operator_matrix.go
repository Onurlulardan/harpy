package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func validateOperator(
	fieldType api.FieldType, kind api.LogicalType, operator api.Operator,
) restqueryv1.IssueCode {
	if isNullOperator(operator) {
		return ""
	}
	allowed := false
	switch kind {
	case api.LogicalTypeText:
		allowed = textOperatorAllowed(fieldType, operator)
	case api.LogicalTypeInteger, api.LogicalTypeDecimal,
		api.LogicalTypeDate, api.LogicalTypeTimestamp:
		allowed = exactOrderOperator(operator)
	case api.LogicalTypeBoolean:
		allowed = equalityListOperator(operator)
	case api.LogicalTypeUUID, api.LogicalTypeRelation:
		allowed = equalityListOperator(operator)
	case api.LogicalTypeEnum:
		allowed = enumOperatorAllowed(fieldType, operator)
	case api.LogicalTypeTextList, api.LogicalTypeUUIDList:
		allowed = collectionOperator(operator)
	case api.LogicalTypeJSON:
		allowed = operator == api.OperatorEq || operator == api.OperatorNe
	}
	if !allowed {
		return restqueryv1.IssueOperatorNotAllowed
	}
	return ""
}

func textOperatorAllowed(fieldType api.FieldType, operator api.Operator) bool {
	if exactOrderOperator(operator) || equalityListOperator(operator) {
		return true
	}
	if !patternOperator(operator) {
		return false
	}
	profile, found, err := fieldType.Text()
	if err != nil || !found {
		return false
	}
	_, enabled, err := profile.Casefold()
	return err == nil && enabled
}

func enumOperatorAllowed(fieldType api.FieldType, operator api.Operator) bool {
	if equalityListOperator(operator) {
		return true
	}
	if !orderOperator(operator) && operator != api.OperatorBetween {
		return false
	}
	profile, found, err := fieldType.Enum()
	if err != nil || !found {
		return false
	}
	ordered, err := profile.Ordered()
	return err == nil && ordered
}

func equalityListOperator(operator api.Operator) bool {
	return operator == api.OperatorEq || operator == api.OperatorNe ||
		operator == api.OperatorIn || operator == api.OperatorNotIn
}

func exactOrderOperator(operator api.Operator) bool {
	return equalityListOperator(operator) || orderOperator(operator) ||
		operator == api.OperatorBetween
}

func orderOperator(operator api.Operator) bool {
	return operator == api.OperatorGT || operator == api.OperatorGTE ||
		operator == api.OperatorLT || operator == api.OperatorLTE
}

func patternOperator(operator api.Operator) bool {
	return operator == api.OperatorContains || operator == api.OperatorStartsWith ||
		operator == api.OperatorEndsWith
}

func collectionOperator(operator api.Operator) bool {
	return operator == api.OperatorContainsAny || operator == api.OperatorContainsAll
}

func isNullOperator(operator api.Operator) bool {
	return operator == api.OperatorIsNull || operator == api.OperatorIsNotNull
}
