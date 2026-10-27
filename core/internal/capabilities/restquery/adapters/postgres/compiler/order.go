package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func appendRecordOrder(
	builder *strings.Builder, query resolved.RecordQuery, binder *binder,
) error {
	if builder == nil || binder == nil {
		return ErrContract
	}
	sorts, err := query.Sorts()
	if err != nil || len(sorts) == 0 {
		return ErrContract
	}
	compiled := make([]string, len(sorts))
	for index, sort := range sorts {
		field, fieldErr := sort.Field()
		direction, directionErr := sort.Direction()
		if fieldErr != nil || directionErr != nil {
			return ErrContract
		}
		expression, expressionErr := orderExpression(field, binder)
		if expressionErr != nil {
			return expressionErr
		}
		suffix := " ASC NULLS LAST"
		if direction == api.DirectionDescending {
			suffix = " DESC NULLS LAST"
		} else if direction != api.DirectionAscending {
			return ErrContract
		}
		compiled[index] = expression + suffix
	}
	builder.WriteString(" ORDER BY ")
	builder.WriteString(strings.Join(compiled, ", "))
	return nil
}

func orderExpression(field api.FieldDefinition, binder *binder) (string, error) {
	_, kind, err := fieldIdentity(field)
	if err != nil {
		return "", err
	}
	if kind == api.LogicalTypeEnum {
		return enumOrderExpression(field, binder)
	}
	return exactExpression(field)
}
