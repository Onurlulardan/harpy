package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

type aggregateInput struct {
	table      sqlidentifier.Token
	query      resolved.AggregateQuery
	groups     []aggregateGroup
	measures   []aggregateMeasure
	visibility resolved.ResourceVisibility
	binding    statementBinding
}

type aggregateGroup struct {
	declaration resolved.AggregateGroup
	field       api.FieldDefinition
	key         string
}

type aggregateMeasure struct {
	declaration resolved.AggregateMeasure
	field       api.FieldDefinition
	key         string
	function    api.AggregateFunction
	hasField    bool
}

func (input aggregateInput) validate() error {
	if input.table.Validate() != nil || input.query.Validate() != nil ||
		input.visibility.Validate() != nil || input.binding.validate() != nil ||
		len(input.groups) > 3 || len(input.measures) > 8 ||
		len(input.groups)+len(input.measures) < 1 {
		return ErrContract
	}
	for _, group := range input.groups {
		if group.declaration.Validate() != nil || group.field.Validate() != nil ||
			!validAggregateOutputKey(group.key, false) {
			return ErrContract
		}
	}
	for _, measure := range input.measures {
		if measure.declaration.Validate() != nil ||
			measure.function.String() == "" ||
			!validAggregateOutputKey(measure.key, true) ||
			measure.hasField != (measure.field.Validate() == nil) ||
			(measure.function == api.AggregateFunctionCount) == measure.hasField {
			return ErrContract
		}
	}
	return nil
}
