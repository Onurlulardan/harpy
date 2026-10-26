package api

import (
	"errors"
	"fmt"
)

var errInvalidCatalogSet = errors.New("REST Query catalog set is invalid")

type CatalogSet struct {
	root      CatalogReference
	catalogs  []FieldCatalog
	validated bool
}

func NewCatalogSet(
	root CatalogReference, catalogs []FieldCatalog,
) (CatalogSet, error) {
	set := CatalogSet{
		root: root, catalogs: cloneFieldCatalogs(catalogs), validated: true,
	}
	if err := set.Validate(); err != nil {
		return CatalogSet{}, err
	}
	return set, nil
}

func (set CatalogSet) Validate() error {
	if !set.validated || set.root.Validate() != nil ||
		set.catalogs == nil || len(set.catalogs) == 0 {
		return errInvalidCatalogSet
	}
	exact, valid := catalogSetIndex(set.catalogs)
	if !valid {
		return errInvalidCatalogSet
	}
	if _, found := exact[exactCatalogKeyOf(set.root)]; !found ||
		!catalogRelationsResolve(set.catalogs, exact) {
		return errInvalidCatalogSet
	}
	return nil
}

func (set CatalogSet) Root() (CatalogReference, error) {
	if err := set.Validate(); err != nil {
		return CatalogReference{}, err
	}
	return set.root, nil
}

func (set CatalogSet) Catalogs() ([]FieldCatalog, error) {
	if err := set.Validate(); err != nil {
		return nil, err
	}
	return cloneFieldCatalogs(set.catalogs), nil
}

func (set CatalogSet) Catalog(
	reference CatalogReference,
) (FieldCatalog, bool, error) {
	if err := set.Validate(); err != nil || reference.Validate() != nil {
		return FieldCatalog{}, false, errInvalidCatalogSet
	}
	wanted := exactCatalogKeyOf(reference)
	for _, catalog := range set.catalogs {
		if exactCatalogKeyOf(catalog.reference) == wanted {
			return catalog.clone(), true, nil
		}
	}
	return FieldCatalog{}, false, nil
}

func (set CatalogSet) RootCatalog() (FieldCatalog, error) {
	catalog, found, err := set.Catalog(set.root)
	if err != nil || !found {
		return FieldCatalog{}, errInvalidCatalogSet
	}
	return catalog, nil
}

func (CatalogSet) String() string   { return queryValueRedaction }
func (CatalogSet) GoString() string { return queryValueRedaction }
func (CatalogSet) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (CatalogSet) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }

func cloneFieldCatalogs(catalogs []FieldCatalog) []FieldCatalog {
	result := make([]FieldCatalog, len(catalogs))
	for index, catalog := range catalogs {
		result[index] = catalog.clone()
	}
	return result
}
