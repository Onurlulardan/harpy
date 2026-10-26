package resolved

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (filter MandatoryFilter) Kind() (api.RequestFilterKind, error) {
	if err := filter.Validate(); err != nil {
		return 0, err
	}
	return filter.kind, nil
}

func (filter MandatoryFilter) Predicate() (MandatoryPredicate, bool, error) {
	if err := filter.Validate(); err != nil {
		return MandatoryPredicate{}, false, err
	}
	if filter.predicate == nil {
		return MandatoryPredicate{}, false, nil
	}
	return filter.predicate.clone(), true, nil
}

func (filter MandatoryFilter) Children() ([]MandatoryFilter, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return cloneMandatoryFilters(filter.children), nil
}

func (predicate MandatoryPredicate) Field() (api.FieldDefinition, error) {
	if err := predicate.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return predicate.field, nil
}

func (predicate MandatoryPredicate) Operator() (api.Operator, error) {
	if err := predicate.Validate(); err != nil {
		return 0, err
	}
	return predicate.operator, nil
}

func (predicate MandatoryPredicate) Values() ([]Value, error) {
	if err := predicate.Validate(); err != nil {
		return nil, err
	}
	return cloneValues(predicate.values), nil
}
