package api

func filterOperatorAllowed(capabilities FieldCapabilities, operator Operator) bool {
	kind := capabilities.logicalType
	switch operator {
	case OperatorIsNull, OperatorIsNotNull:
		return true
	case OperatorContainsAny, OperatorContainsAll:
		return kind == LogicalTypeTextList || kind == LogicalTypeUUIDList
	case OperatorContains, OperatorStartsWith, OperatorEndsWith:
		return kind == LogicalTypeText && capabilities.casefoldText
	case OperatorGT, OperatorGTE, OperatorLT, OperatorLTE, OperatorBetween:
		return orderedFilterType(kind, capabilities.orderedEnum)
	case OperatorIn, OperatorNotIn:
		return scalarLogicalType(kind) && kind != LogicalTypeJSON
	case OperatorEq, OperatorNe:
		return scalarLogicalType(kind) || kind == LogicalTypeJSON
	default:
		return false
	}
}

func orderedFilterType(kind LogicalType, orderedEnum bool) bool {
	switch kind {
	case LogicalTypeText, LogicalTypeInteger, LogicalTypeDecimal,
		LogicalTypeDate, LogicalTypeTimestamp:
		return true
	case LogicalTypeEnum:
		return orderedEnum
	default:
		return false
	}
}

func scalarLogicalType(kind LogicalType) bool {
	return kind >= LogicalTypeText && kind <= LogicalTypeRelation
}

func sortableLogicalType(kind LogicalType, orderedEnum bool) bool {
	return scalarLogicalType(kind) && (kind != LogicalTypeEnum || orderedEnum)
}

func groupableLogicalType(kind LogicalType) bool {
	return scalarLogicalType(kind)
}

func aggregateFunctionAllowed(
	capabilities FieldCapabilities, function AggregateFunction,
) bool {
	kind := capabilities.logicalType
	switch function {
	case AggregateFunctionCountDistinct:
		return scalarLogicalType(kind)
	case AggregateFunctionSum, AggregateFunctionAvg:
		return kind == LogicalTypeInteger || kind == LogicalTypeDecimal
	case AggregateFunctionMin, AggregateFunctionMax:
		return aggregateOrderedType(kind, capabilities.orderedEnum)
	default:
		return false
	}
}

func aggregateOrderedType(kind LogicalType, orderedEnum bool) bool {
	switch kind {
	case LogicalTypeText, LogicalTypeInteger, LogicalTypeDecimal,
		LogicalTypeDate, LogicalTypeTimestamp:
		return true
	case LogicalTypeEnum:
		return orderedEnum
	default:
		return false
	}
}

func validCursorCapability(capabilities FieldCapabilities) bool {
	if !capabilities.sortable {
		return capabilities.cursorCanonicalBytes == 0
	}
	return capabilities.cursorCanonicalBytes >= 1 &&
		capabilities.cursorCanonicalBytes <= maximumCursorCanonicalBytes
}
