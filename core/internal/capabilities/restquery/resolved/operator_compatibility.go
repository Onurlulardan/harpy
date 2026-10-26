package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func operatorAllowedForType(fieldType api.FieldType, operator api.Operator) bool {
	if operator == api.OperatorIsNull || operator == api.OperatorIsNotNull {
		return true
	}
	kind, err := fieldType.Kind()
	if err != nil {
		return false
	}
	switch kind {
	case api.LogicalTypeText:
		return textOperatorAllowed(fieldType, operator)
	case api.LogicalTypeInteger, api.LogicalTypeDecimal,
		api.LogicalTypeDate, api.LogicalTypeTimestamp:
		return exactOrderOperator(operator)
	case api.LogicalTypeBoolean, api.LogicalTypeUUID, api.LogicalTypeRelation:
		return equalityListOperator(operator)
	case api.LogicalTypeEnum:
		return enumOperatorAllowed(fieldType, operator)
	case api.LogicalTypeTextList, api.LogicalTypeUUIDList:
		return operator == api.OperatorContainsAny || operator == api.OperatorContainsAll
	case api.LogicalTypeJSON:
		return operator == api.OperatorEq || operator == api.OperatorNe
	default:
		return false
	}
}

func textOperatorAllowed(fieldType api.FieldType, operator api.Operator) bool {
	if exactOrderOperator(operator) {
		return true
	}
	if operator != api.OperatorContains && operator != api.OperatorStartsWith &&
		operator != api.OperatorEndsWith {
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
	if operator != api.OperatorGT && operator != api.OperatorGTE &&
		operator != api.OperatorLT && operator != api.OperatorLTE &&
		operator != api.OperatorBetween {
		return false
	}
	profile, found, err := fieldType.Enum()
	if err != nil || !found {
		return false
	}
	ordered, err := profile.Ordered()
	return err == nil && ordered
}

func exactOrderOperator(operator api.Operator) bool {
	return equalityListOperator(operator) || operator == api.OperatorGT ||
		operator == api.OperatorGTE || operator == api.OperatorLT ||
		operator == api.OperatorLTE || operator == api.OperatorBetween
}

func equalityListOperator(operator api.Operator) bool {
	return operator == api.OperatorEq || operator == api.OperatorNe ||
		operator == api.OperatorIn || operator == api.OperatorNotIn
}
