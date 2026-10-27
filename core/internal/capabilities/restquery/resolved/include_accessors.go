package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func (include Include) Relation() (api.RelationDefinition, error) {
	if err := include.Validate(); err != nil {
		return api.RelationDefinition{}, err
	}
	return include.relation, nil
}

func (include Include) Target() (api.CatalogReference, error) {
	if err := include.Validate(); err != nil {
		return api.CatalogReference{}, err
	}
	return include.target, nil
}

func (include Include) TargetStorage() (api.ResourceStorage, error) {
	if err := include.Validate(); err != nil {
		return api.ResourceStorage{}, err
	}
	return include.targetStorage, nil
}

func (include Include) RelationStorage() (api.RelationStorage, error) {
	if err := include.Validate(); err != nil {
		return api.RelationStorage{}, err
	}
	return include.relationStorage, nil
}

func (include Include) StableID() (api.FieldDefinition, error) {
	if err := include.Validate(); err != nil {
		return api.FieldDefinition{}, err
	}
	return include.stableID, nil
}

func (include Include) Path() (RequestPath, error) {
	if err := include.Validate(); err != nil {
		return RequestPath{}, err
	}
	return include.path, nil
}

func (include Include) Projections() ([]Projection, error) {
	if err := include.Validate(); err != nil {
		return nil, err
	}
	return append([]Projection{}, include.projections...), nil
}

func (include Include) Where() (Filter, bool, error) {
	if err := include.Validate(); err != nil {
		return Filter{}, false, err
	}
	if include.where == nil {
		return Filter{}, false, nil
	}
	return include.where.clone(), true, nil
}

func (include Include) Sorts() ([]Sort, error) {
	if err := include.Validate(); err != nil {
		return nil, err
	}
	return append([]Sort{}, include.sorts...), nil
}

func (include Include) Limit() (int, error) {
	if err := include.Validate(); err != nil {
		return 0, err
	}
	return include.limit, nil
}

func (include Include) Children() ([]Include, error) {
	if err := include.Validate(); err != nil {
		return nil, err
	}
	return cloneIncludes(include.children), nil
}
