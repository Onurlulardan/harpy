package resolved

func (filter MandatoryFilter) clone() MandatoryFilter {
	result := filter
	result.children = cloneMandatoryFilters(filter.children)
	if filter.predicate != nil {
		owned := filter.predicate.clone()
		result.predicate = &owned
	}
	return result
}

func cloneMandatoryFilters(values []MandatoryFilter) []MandatoryFilter {
	if values == nil {
		return nil
	}
	result := make([]MandatoryFilter, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}

func (predicate MandatoryPredicate) clone() MandatoryPredicate {
	result := predicate
	result.values = cloneValues(predicate.values)
	return result
}
