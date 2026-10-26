package api

func (capabilities FieldCapabilities) LogicalType() (LogicalType, error) {
	if err := capabilities.Validate(); err != nil {
		return 0, err
	}
	return capabilities.logicalType, nil
}

func (capabilities FieldCapabilities) Selectable() (bool, error) {
	if err := capabilities.Validate(); err != nil {
		return false, err
	}
	return capabilities.selectable, nil
}

func (capabilities FieldCapabilities) Filters() ([]FilterCapability, error) {
	if err := capabilities.Validate(); err != nil {
		return nil, err
	}
	return append([]FilterCapability{}, capabilities.filters...), nil
}

func (capabilities FieldCapabilities) Filter(
	operator Operator,
) (FilterCapability, bool, error) {
	if err := capabilities.Validate(); err != nil {
		return FilterCapability{}, false, err
	}
	for _, capability := range capabilities.filters {
		if capability.operator == operator {
			return capability, true, nil
		}
	}
	return FilterCapability{}, false, nil
}

func (capabilities FieldCapabilities) Sortable() (bool, error) {
	if err := capabilities.Validate(); err != nil {
		return false, err
	}
	return capabilities.sortable, nil
}

func (capabilities FieldCapabilities) CursorCanonicalBytes() (int, error) {
	if err := capabilities.Validate(); err != nil {
		return 0, err
	}
	return capabilities.cursorCanonicalBytes, nil
}

func (capabilities FieldCapabilities) Searchable() (bool, error) {
	if err := capabilities.Validate(); err != nil {
		return false, err
	}
	return capabilities.searchable, nil
}

func (capabilities FieldCapabilities) Groupable() (bool, error) {
	if err := capabilities.Validate(); err != nil {
		return false, err
	}
	return capabilities.groupable, nil
}

func (capabilities FieldCapabilities) Aggregates() ([]AggregateFunction, error) {
	if err := capabilities.Validate(); err != nil {
		return nil, err
	}
	return append([]AggregateFunction{}, capabilities.aggregates...), nil
}

func (capabilities FieldCapabilities) AllowsAggregate(
	function AggregateFunction,
) (bool, error) {
	if err := capabilities.Validate(); err != nil {
		return false, err
	}
	for _, allowed := range capabilities.aggregates {
		if allowed == function {
			return true, nil
		}
	}
	return false, nil
}

func (capabilities FieldCapabilities) clone() FieldCapabilities {
	result := capabilities
	result.filters = append([]FilterCapability{}, capabilities.filters...)
	result.aggregates = append([]AggregateFunction{}, capabilities.aggregates...)
	return result
}
