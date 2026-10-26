package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func appendAggregateOrder(
	builder *strings.Builder, input aggregateInput, binder *binder,
) error {
	if builder == nil || binder == nil || input.validate() != nil {
		return ErrContract
	}
	sorts, err := input.query.Sorts()
	if err != nil {
		return ErrContract
	}
	if len(sorts) == 0 {
		if len(input.groups) != 0 {
			return ErrContract
		}
		return nil
	}
	compiled := make([]string, len(sorts))
	for index, sort := range sorts {
		reference, referenceErr := sort.Reference()
		direction, directionErr := sort.Direction()
		target, targetErr := aggregateReferenceTarget(input, reference)
		if referenceErr != nil || directionErr != nil || targetErr != nil {
			return ErrContract
		}
		expression, err := aggregateTargetExpression(target, true, binder)
		if err != nil {
			return err
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
