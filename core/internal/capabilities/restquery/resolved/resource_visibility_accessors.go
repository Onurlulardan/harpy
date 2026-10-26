package resolved

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func (visibility ResourceVisibility) RequirementSHA256() ([sha256.Size]byte, error) {
	if err := visibility.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return visibility.requirement, nil
}

func (visibility ResourceVisibility) Catalog() (api.CatalogReference, error) {
	if err := visibility.Validate(); err != nil {
		return api.CatalogReference{}, err
	}
	return visibility.catalog, nil
}

func (visibility ResourceVisibility) Path() (RequestPath, error) {
	if err := visibility.Validate(); err != nil {
		return RequestPath{}, err
	}
	return visibility.path, nil
}

func (visibility ResourceVisibility) Mode() (ResourceVisibilityMode, error) {
	if err := visibility.Validate(); err != nil {
		return 0, err
	}
	return visibility.mode, nil
}

func (visibility ResourceVisibility) Filter() (MandatoryFilter, bool, error) {
	if err := visibility.Validate(); err != nil {
		return MandatoryFilter{}, false, err
	}
	if visibility.filter == nil {
		return MandatoryFilter{}, false, nil
	}
	return visibility.filter.clone(), true, nil
}

func (visibility ResourceVisibility) clone() ResourceVisibility {
	result := visibility
	if visibility.filter != nil {
		owned := visibility.filter.clone()
		result.filter = &owned
	}
	return result
}

func cloneResourceVisibilities(values []ResourceVisibility) []ResourceVisibility {
	result := make([]ResourceVisibility, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}
