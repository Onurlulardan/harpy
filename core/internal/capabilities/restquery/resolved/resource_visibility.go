package resolved

import (
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type ResourceVisibilityMode uint8

const (
	ResourceVisibilityAll ResourceVisibilityMode = iota + 1
	ResourceVisibilityNone
	ResourceVisibilityPredicate
)

type ResourceVisibility struct {
	requirement [sha256.Size]byte
	catalog     api.CatalogReference
	path        RequestPath
	mode        ResourceVisibilityMode
	filter      *MandatoryFilter
	validated   bool
}

func NewAllResourceVisibility(
	requirement [sha256.Size]byte,
	catalog api.CatalogReference,
	path RequestPath,
) (ResourceVisibility, error) {
	return newResourceVisibility(requirement, catalog, path, ResourceVisibilityAll, nil)
}

func NewNoResourceVisibility(
	requirement [sha256.Size]byte,
	catalog api.CatalogReference,
	path RequestPath,
) (ResourceVisibility, error) {
	return newResourceVisibility(requirement, catalog, path, ResourceVisibilityNone, nil)
}

func NewPredicateResourceVisibility(
	requirement [sha256.Size]byte,
	catalog api.CatalogReference,
	path RequestPath,
	filter MandatoryFilter,
) (ResourceVisibility, error) {
	return newResourceVisibility(
		requirement, catalog, path, ResourceVisibilityPredicate, &filter,
	)
}

func newResourceVisibility(
	requirement [sha256.Size]byte,
	catalog api.CatalogReference,
	path RequestPath,
	mode ResourceVisibilityMode,
	filter *MandatoryFilter,
) (ResourceVisibility, error) {
	visibility := ResourceVisibility{
		requirement: requirement, catalog: catalog, path: path,
		mode: mode, validated: true,
	}
	if filter != nil {
		owned := filter.clone()
		visibility.filter = &owned
	}
	if err := visibility.Validate(); err != nil {
		return ResourceVisibility{}, err
	}
	return visibility, nil
}

func (visibility ResourceVisibility) Validate() error {
	if !visibility.validated || visibility.requirement == [sha256.Size]byte{} ||
		visibility.catalog.Validate() != nil || visibility.path.Validate() != nil ||
		visibility.mode.Validate() != nil {
		return fmt.Errorf("resolved REST Query resource visibility is invalid")
	}
	if visibility.mode == ResourceVisibilityPredicate {
		if visibility.filter == nil || visibility.filter.Validate() != nil {
			return fmt.Errorf("resolved REST Query resource visibility is invalid")
		}
	} else if visibility.filter != nil {
		return fmt.Errorf("resolved REST Query resource visibility is invalid")
	}
	return nil
}

func (mode ResourceVisibilityMode) Validate() error {
	if mode.String() == "" {
		return fmt.Errorf("resolved REST Query visibility mode is invalid")
	}
	return nil
}

func (mode ResourceVisibilityMode) String() string {
	switch mode {
	case ResourceVisibilityAll:
		return "all"
	case ResourceVisibilityNone:
		return "none"
	case ResourceVisibilityPredicate:
		return "predicate"
	default:
		return ""
	}
}

func (ResourceVisibility) String() string   { return valueRedaction }
func (ResourceVisibility) GoString() string { return valueRedaction }
func (visibility ResourceVisibility) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (ResourceVisibility) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
func (ResourceVisibilityMode) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
