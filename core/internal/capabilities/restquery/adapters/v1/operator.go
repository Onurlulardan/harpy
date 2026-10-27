package v1query

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type legacyOperation uint8

const (
	legacyDirect legacyOperation = iota + 1
	legacyNotContains
	legacyIsEmpty
	legacyIsNotEmpty
)

type resolvedOperation struct {
	kind     legacyOperation
	operator api.Operator
}

func legacyOperator(value string) (resolvedOperation, bool) {
	switch value {
	case "equals":
		return directOperation(api.OperatorEq), true
	case "not_equals":
		return directOperation(api.OperatorNe), true
	case "contains":
		return directOperation(api.OperatorContains), true
	case "starts_with":
		return directOperation(api.OperatorStartsWith), true
	case "ends_with":
		return directOperation(api.OperatorEndsWith), true
	case "greater_than":
		return directOperation(api.OperatorGT), true
	case "less_than":
		return directOperation(api.OperatorLT), true
	case "greater_than_or_equal":
		return directOperation(api.OperatorGTE), true
	case "less_than_or_equal":
		return directOperation(api.OperatorLTE), true
	case "in":
		return directOperation(api.OperatorIn), true
	case "not_in":
		return directOperation(api.OperatorNotIn), true
	case "array_contains":
		return directOperation(api.OperatorContainsAny), true
	case "not_contains":
		return resolvedOperation{kind: legacyNotContains}, true
	case "is_empty":
		return resolvedOperation{kind: legacyIsEmpty}, true
	case "is_not_empty":
		return resolvedOperation{kind: legacyIsNotEmpty}, true
	default:
		return resolvedOperation{}, false
	}
}

func directOperation(operator api.Operator) resolvedOperation {
	return resolvedOperation{kind: legacyDirect, operator: operator}
}

func operationMatchesShape(operation resolvedOperation, shape FilterShape) bool {
	collectionOperator := operation.kind == legacyDirect &&
		(operation.operator == api.OperatorContainsAny ||
			operation.operator == api.OperatorContainsAll)
	if shape != FilterShapeCollection {
		return !collectionOperator
	}
	return collectionOperator || operation.kind == legacyDirect &&
		(operation.operator == api.OperatorIsNull ||
			operation.operator == api.OperatorIsNotNull)
}

func resolveOperator(
	raw string, path string, policy manifest, aliases *aliasSet,
) (resolvedOperation, bool) {
	if operation, found := legacyOperator(raw); found {
		return operation, true
	}
	binding, found := policy.operatorAliases[raw]
	if !found {
		return resolvedOperation{}, false
	}
	aliases.add(binding.aliasID, path, restqueryv1.IssueOperatorUnknown)
	return resolvedOperation{kind: legacyDirect, operator: binding.operator}, true
}
