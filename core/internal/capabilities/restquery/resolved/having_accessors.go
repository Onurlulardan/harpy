package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func (having Having) Kind() (api.RequestHavingKind, error) {
	if err := having.Validate(); err != nil {
		return 0, err
	}
	return having.kind, nil
}

func (having Having) Predicate() (HavingPredicate, bool, error) {
	if err := having.Validate(); err != nil {
		return HavingPredicate{}, false, err
	}
	if having.predicate == nil {
		return HavingPredicate{}, false, nil
	}
	return having.predicate.clone(), true, nil
}

func (having Having) Children() ([]Having, error) {
	if err := having.Validate(); err != nil {
		return nil, err
	}
	return cloneHavings(having.children), nil
}

func (predicate HavingPredicate) Reference() (AggregateReference, error) {
	if err := predicate.Validate(); err != nil {
		return AggregateReference{}, err
	}
	return predicate.reference, nil
}

func (predicate HavingPredicate) Operator() (api.Operator, error) {
	if err := predicate.Validate(); err != nil {
		return 0, err
	}
	return predicate.operator, nil
}

func (predicate HavingPredicate) Values() ([]Value, error) {
	if err := predicate.Validate(); err != nil {
		return nil, err
	}
	return cloneValues(predicate.values), nil
}
