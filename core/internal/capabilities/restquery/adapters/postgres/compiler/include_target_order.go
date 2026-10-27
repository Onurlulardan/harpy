package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func appendIncludeTargetOrder(
	builder *strings.Builder, sorts []resolved.Sort, binder *binder,
) error {
	if builder == nil || binder == nil || len(sorts) == 0 {
		return ErrContract
	}
	compiled := make([]string, len(sorts))
	for index, sort := range sorts {
		field, fieldErr := sort.Field()
		direction, directionErr := sort.Direction()
		if fieldErr != nil || directionErr != nil {
			return ErrContract
		}
		expression, err := orderExpression(field, binder)
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
