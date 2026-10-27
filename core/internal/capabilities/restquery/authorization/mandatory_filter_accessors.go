package authorization

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
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

func (predicate MandatoryPredicate) FieldID() (uuid.UUID, error) {
	if err := predicate.Validate(); err != nil {
		return uuid.Nil, err
	}
	return predicate.fieldID, nil
}

func (predicate MandatoryPredicate) LogicalType() (api.LogicalType, error) {
	if err := predicate.Validate(); err != nil {
		return 0, err
	}
	return predicate.logicalType, nil
}

func (predicate MandatoryPredicate) Operator() (api.Operator, error) {
	if err := predicate.Validate(); err != nil {
		return 0, err
	}
	return predicate.operator, nil
}

func (predicate MandatoryPredicate) Values() ([]resolved.Value, error) {
	if err := predicate.Validate(); err != nil {
		return nil, err
	}
	return append([]resolved.Value{}, predicate.values...), nil
}
