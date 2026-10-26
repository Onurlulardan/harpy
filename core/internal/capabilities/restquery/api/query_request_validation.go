package api

import "fmt"

func (request QueryRequest) Validate() error {
	switch request.mode {
	case QueryModeRecord:
		return request.validateRecord()
	case QueryModeAggregate:
		return request.validateAggregate()
	default:
		return fmt.Errorf("REST Query mode is invalid")
	}
}

func (request QueryRequest) validateRecord() error {
	if request.aggregate != nil {
		return fmt.Errorf("REST Query record mode contains aggregate state")
	}
	if request.selectFields == nil || request.sorts == nil || request.includes == nil {
		return fmt.Errorf("REST Query record collections are not canonical")
	}
	if _, err := cloneUniqueRequestNames(request.selectFields, maximumProjectedFields); err != nil {
		return fmt.Errorf("REST Query root select: %w", err)
	}
	if _, err := cloneRequestSorts(request.sorts); err != nil {
		return err
	}
	if request.where != nil && !request.where.valid() {
		return fmt.Errorf("REST Query filter is invalid")
	}
	if !request.page.valid() || !request.count.valid() {
		return fmt.Errorf("REST Query record mode is invalid")
	}
	if request.search != nil && !request.search.valid() {
		return fmt.Errorf("REST Query search is invalid")
	}
	includes, err := cloneRequestIncludes(request.includes)
	if err != nil {
		return err
	}
	projected := len(request.selectFields)
	for _, include := range includes {
		projected += int(include.projectedFields)
	}
	if projected > maximumProjectedFields {
		return fmt.Errorf("REST Query projected field count exceeds its limit")
	}
	return validateQueryFilterBudget(request.where, request.includes)
}

func (request QueryRequest) validateAggregate() error {
	if request.aggregate == nil || request.selectFields != nil || request.sorts != nil ||
		request.includes != nil || request.page != (RequestPage{}) || request.count != 0 {
		return fmt.Errorf("REST Query aggregate mode contains record state")
	}
	if request.where != nil && !request.where.valid() {
		return fmt.Errorf("REST Query filter is invalid")
	}
	if request.search != nil && !request.search.valid() {
		return fmt.Errorf("REST Query search is invalid")
	}
	if err := request.aggregate.Validate(); err != nil {
		return err
	}
	return validateAggregateQueryBudget(request.where, request.aggregate.having)
}

func cloneUniqueRequestNames(values []string, maximum int) ([]string, error) {
	if len(values) > maximum {
		return nil, fmt.Errorf("field count exceeds its limit")
	}
	result := append([]string{}, values...)
	seen := make(map[string]struct{}, len(result))
	for _, value := range result {
		if !validRequestName(value) {
			return nil, fmt.Errorf("field name is invalid")
		}
		if _, duplicate := seen[value]; duplicate {
			return nil, fmt.Errorf("field name is duplicated")
		}
		seen[value] = struct{}{}
	}
	return result, nil
}

func cloneOptionalRequestFilter(value *RequestFilter) (*RequestFilter, error) {
	if value == nil {
		return nil, nil
	}
	if !value.valid() {
		return nil, fmt.Errorf("REST Query filter is invalid")
	}
	result := value.clone()
	return &result, nil
}

func cloneOptionalRequestSearch(value *RequestSearch) (*RequestSearch, error) {
	if value == nil {
		return nil, nil
	}
	if !value.valid() {
		return nil, fmt.Errorf("REST Query search is invalid")
	}
	result := *value
	return &result, nil
}
