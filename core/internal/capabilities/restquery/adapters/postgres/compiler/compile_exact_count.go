package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func compileExactCount(input recordInput) (ExactCountStatement, error) {
	if input.validate() != nil ||
		validateRecordCountShape(input.query, api.CountModeExact) != nil {
		return ExactCountStatement{}, ErrContract
	}
	binder := newBinder()
	conditions, err := rootConditions(input.query, input.visibility, binder)
	if err != nil {
		return ExactCountStatement{}, err
	}
	var builder strings.Builder
	builder.WriteString("SELECT COUNT(*) FROM ")
	if err := appendResourceStorage(&builder, input.storage); err != nil {
		return ExactCountStatement{}, err
	}
	if len(conditions) != 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}
	return newExactCountStatement(
		builder.String(), binder.arguments,
		planner.StatementClassExactCount, input.binding,
	)
}
