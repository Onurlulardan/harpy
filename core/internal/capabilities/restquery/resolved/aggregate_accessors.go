package resolved

func (query AggregateQuery) Groups() ([]AggregateGroup, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return append([]AggregateGroup{}, query.groups...), nil
}

func (query AggregateQuery) Measures() ([]AggregateMeasure, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return append([]AggregateMeasure{}, query.measures...), nil
}

func (query AggregateQuery) Where() (Filter, bool, error) {
	if err := query.Validate(); err != nil {
		return Filter{}, false, err
	}
	if query.where == nil {
		return Filter{}, false, nil
	}
	return query.where.clone(), true, nil
}

func (query AggregateQuery) Search() ([]SearchTerm, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return cloneSearchTerms(query.search), nil
}

func (query AggregateQuery) Having() (Having, bool, error) {
	if err := query.Validate(); err != nil {
		return Having{}, false, err
	}
	if query.having == nil {
		return Having{}, false, nil
	}
	return query.having.clone(), true, nil
}

func (query AggregateQuery) Sorts() ([]AggregateSort, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return append([]AggregateSort{}, query.sorts...), nil
}

func (query AggregateQuery) Limit() (int, error) {
	if err := query.Validate(); err != nil {
		return 0, err
	}
	return query.limit, nil
}
