package api

import "fmt"

func (aggregate RequestAggregate) Validate() error {
	if aggregate.groupBy == nil || aggregate.measures == nil || aggregate.sorts == nil {
		return fmt.Errorf("REST Query aggregate collections are not canonical")
	}
	if len(aggregate.groupBy) == 0 && len(aggregate.measures) == 0 {
		return fmt.Errorf("REST Query aggregate declaration is empty")
	}
	groupBy, err := cloneAggregateGroups(aggregate.groupBy)
	if err != nil {
		return err
	}
	measures, err := cloneRequestMeasures(aggregate.measures)
	if err != nil {
		return err
	}
	if aggregate.limit < 1 || aggregate.limit > maximumAggregateLimit {
		return fmt.Errorf("REST Query aggregate limit is invalid")
	}
	groups, measureNames := aggregateDeclarations(groupBy, measures)
	if aggregate.having != nil &&
		(!aggregate.having.valid() ||
			!aggregate.having.referencesDeclared(groups, measureNames)) {
		return fmt.Errorf("REST Query aggregate HAVING is invalid")
	}
	if _, err := cloneRequestAggregateSorts(
		aggregate.sorts, groups, measureNames,
	); err != nil {
		return err
	}
	return nil
}

func cloneAggregateGroups(values []string) ([]string, error) {
	groups, err := cloneUniqueRequestNames(values, maximumAggregateGroups)
	if err != nil {
		return nil, fmt.Errorf("REST Query aggregate group: %w", err)
	}
	return groups, nil
}

func aggregateDeclarations(
	groups []string, measures []RequestMeasure,
) (map[string]struct{}, map[string]struct{}) {
	groupSet := make(map[string]struct{}, len(groups))
	for _, group := range groups {
		groupSet[group] = struct{}{}
	}
	measureSet := make(map[string]struct{}, len(measures))
	for _, measure := range measures {
		measureSet[measure.name] = struct{}{}
	}
	return groupSet, measureSet
}

func cloneOptionalRequestHaving(
	value *RequestHaving,
	groups map[string]struct{},
	measures map[string]struct{},
) (*RequestHaving, error) {
	if value == nil {
		return nil, nil
	}
	if !value.valid() || !value.referencesDeclared(groups, measures) {
		return nil, fmt.Errorf("REST Query aggregate HAVING is invalid")
	}
	result := value.clone()
	return &result, nil
}
