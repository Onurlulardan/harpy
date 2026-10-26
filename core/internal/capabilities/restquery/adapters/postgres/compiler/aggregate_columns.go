package compiler

func aggregateColumns(
	input aggregateInput,
) ([]AggregateGroupColumn, []AggregateMeasureColumn, error) {
	if input.validate() != nil {
		return nil, nil, ErrContract
	}
	groups := make([]AggregateGroupColumn, len(input.groups))
	for index, group := range input.groups {
		column, err := newAggregateGroupColumn(group.field, group.key)
		if err != nil {
			return nil, nil, ErrContract
		}
		groups[index] = column
	}
	measures := make([]AggregateMeasureColumn, len(input.measures))
	for index, measure := range input.measures {
		column, err := newAggregateMeasureColumn(measure)
		if err != nil {
			return nil, nil, ErrContract
		}
		measures[index] = column
	}
	return groups, measures, nil
}
