package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func (compiler Compiler) extractAggregate(
	plan resolved.AuthorizedPlan,
) (aggregateInput, error) {
	if compiler.Validate() != nil || plan.Validate() != nil {
		return aggregateInput{}, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return aggregateInput{}, ErrContract
	}
	mode, err := spec.Mode()
	query, found, queryErr := spec.Aggregate()
	if err != nil || mode != api.QueryModeAggregate || queryErr != nil || !found {
		return aggregateInput{}, ErrContract
	}
	storage, err := spec.Storage()
	if err != nil {
		return aggregateInput{}, ErrContract
	}
	table, err := storage.Table()
	if err != nil {
		return aggregateInput{}, ErrContract
	}
	reference, referenceErr := spec.Reference()
	rootPath, pathErr := spec.RootPath()
	pathValue, valueErr := rootPath.Value()
	if referenceErr != nil || pathErr != nil || valueErr != nil {
		return aggregateInput{}, ErrContract
	}
	visibility, err := rootVisibility(plan, reference, pathValue)
	if err != nil {
		return aggregateInput{}, err
	}
	groups, groupErr := extractAggregateGroups(query)
	measures, measureErr := extractAggregateMeasures(query)
	binding, bindingErr := aggregatePlanBinding(plan)
	if groupErr != nil || measureErr != nil || bindingErr != nil {
		return aggregateInput{}, ErrContract
	}
	input := aggregateInput{
		table: table, query: query, groups: groups, measures: measures,
		visibility: visibility, binding: binding,
	}
	if input.validate() != nil {
		return aggregateInput{}, ErrContract
	}
	return input, nil
}
