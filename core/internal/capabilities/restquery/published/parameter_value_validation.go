package published

import (
	"bytes"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func validateResolvedParameterValues(
	fieldType api.FieldType, arity ParameterArity, values []resolved.Value,
) bool {
	if arity == ParameterAritySingle && len(values) != 1 ||
		arity == ParameterArityList && (len(values) < 1 || len(values) > 100) {
		return false
	}
	literals := make([]api.Literal, 0, len(values))
	for _, value := range values {
		literal, valid := literalFromResolvedValue(value)
		if !valid {
			return false
		}
		literals = append(literals, literal)
	}
	operator := api.OperatorEq
	if arity == ParameterArityList {
		operator = api.OperatorIn
	}
	normalized, err := resolver.ResolveLiterals(fieldType, operator, literals)
	if err != nil || len(normalized) != len(values) {
		return false
	}
	for index := range values {
		if !resolvedValuesEqual(values[index], normalized[index]) {
			return false
		}
	}
	return true
}

func literalFromResolvedValue(value resolved.Value) (api.Literal, bool) {
	if value.Validate() != nil {
		return api.Literal{}, false
	}
	if scalar, found, err := value.Scalar(); err == nil && found {
		literal, literalErr := api.NewTextLiteral(scalar)
		return literal, literalErr == nil
	}
	if boolean, found, err := value.Boolean(); err == nil && found {
		literal, literalErr := api.NewTextLiteral(strconv.FormatBool(boolean))
		return literal, literalErr == nil
	}
	if document, found, err := value.JSON(); err == nil && found {
		literal, literalErr := api.NewJSONLiteral(document)
		return literal, literalErr == nil
	}
	return api.Literal{}, false
}

func resolvedValuesEqual(left, right resolved.Value) bool {
	leftKind, leftErr := left.Kind()
	rightKind, rightErr := right.Kind()
	if leftErr != nil || rightErr != nil || leftKind != rightKind {
		return false
	}
	if leftScalar, leftFound, _ := left.Scalar(); leftFound {
		rightScalar, rightFound, err := right.Scalar()
		return err == nil && rightFound && leftScalar == rightScalar
	}
	if leftBoolean, leftFound, _ := left.Boolean(); leftFound {
		rightBoolean, rightFound, err := right.Boolean()
		return err == nil && rightFound && leftBoolean == rightBoolean
	}
	leftJSON, leftFound, leftErr := left.JSON()
	rightJSON, rightFound, rightErr := right.JSON()
	return leftErr == nil && rightErr == nil && leftFound && rightFound &&
		bytes.Equal(leftJSON, rightJSON)
}

func cloneResolvedValues(values []resolved.Value) []resolved.Value {
	if values == nil {
		return nil
	}
	return append([]resolved.Value{}, values...)
}
