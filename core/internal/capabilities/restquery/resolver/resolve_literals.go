package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

// ResolveLiterals assigns catalog type semantics to transport-neutral literals.
func ResolveLiterals(
	fieldType api.FieldType, operator api.Operator, literals []api.Literal,
) ([]resolved.Value, error) {
	kind, err := fieldType.Kind()
	if err != nil {
		return nil, literalResolutionFailure(restqueryv1.IssueTypeMismatch)
	}
	arity, validOperator := operator.Arity()
	if !validOperator {
		return nil, literalResolutionFailure(restqueryv1.IssueOperatorNotAllowed)
	}
	if issue := validateOperator(fieldType, kind, operator); issue != "" {
		return nil, literalResolutionFailure(issue)
	}
	if !validLiteralCount(arity, len(literals)) {
		return nil, literalResolutionFailure(restqueryv1.IssueArityMismatch)
	}
	values := make([]resolved.Value, 0, len(literals))
	for _, literal := range literals {
		value, issue := resolveLiteral(fieldType, kind, operator, literal)
		if issue != "" {
			return nil, literalResolutionFailure(issue)
		}
		values = append(values, value)
	}
	if arity == api.PredicateArityList {
		values, err = deduplicateValues(values)
		if err != nil {
			return nil, literalResolutionFailure(restqueryv1.IssueValueNotCanonical)
		}
	}
	return values, nil
}

func validLiteralCount(arity api.PredicateArity, count int) bool {
	switch arity {
	case api.PredicateArityZero:
		return count == 0
	case api.PredicateArityOne:
		return count == 1
	case api.PredicateArityList:
		return count > 0
	case api.PredicateArityPair:
		return count == 2
	default:
		return false
	}
}

func resolveLiteral(
	fieldType api.FieldType,
	kind api.LogicalType,
	operator api.Operator,
	literal api.Literal,
) (resolved.Value, restqueryv1.IssueCode) {
	source, issue := readLiteralSource(literal)
	if issue != "" {
		return resolved.Value{}, issue
	}
	switch kind {
	case api.LogicalTypeText, api.LogicalTypeTextList:
		return resolveText(fieldType, operator, source)
	case api.LogicalTypeInteger:
		return resolveInteger(fieldType, source)
	case api.LogicalTypeDecimal:
		return resolveDecimal(fieldType, source)
	case api.LogicalTypeBoolean:
		return resolveBoolean(source)
	case api.LogicalTypeDate, api.LogicalTypeTimestamp:
		return resolveTemporal(kind, source)
	case api.LogicalTypeUUID, api.LogicalTypeRelation, api.LogicalTypeUUIDList:
		return resolveUUID(kind, source)
	case api.LogicalTypeEnum:
		return resolveEnum(fieldType, source)
	case api.LogicalTypeJSON:
		return resolveJSON(source)
	default:
		return resolved.Value{}, restqueryv1.IssueTypeMismatch
	}
}
