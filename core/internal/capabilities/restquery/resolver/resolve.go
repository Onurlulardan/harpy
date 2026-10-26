package resolver

import (
	"errors"
	"fmt"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

// Resolve binds a validated request to one exact immutable catalog set.
func Resolve(
	request api.QueryRequest, catalogs api.CatalogSet,
) (resolved.QuerySpec, error) {
	if request.Validate() != nil || catalogs.Validate() != nil {
		return resolved.QuerySpec{}, invalidResolutionInput()
	}
	root, err := catalogs.RootCatalog()
	if err != nil {
		return resolved.QuerySpec{}, invalidResolutionInput()
	}
	reference, storage, stableID, err := catalogRootParts(root)
	if err != nil {
		return resolved.QuerySpec{}, invalidResolutionInput()
	}
	collector := newResolutionIssues()
	var spec resolved.QuerySpec
	if request.Mode() == api.QueryModeRecord {
		record := resolveRecord(request, root, catalogs, collector)
		if collector.hasIssues() {
			return resolved.QuerySpec{}, collector.err()
		}
		spec, err = resolved.NewRecordQuerySpec(
			catalogs, reference, storage, stableID, mustRequestPath(""), record,
		)
	} else {
		aggregate := resolveAggregate(request, root, collector)
		if collector.hasIssues() {
			return resolved.QuerySpec{}, collector.err()
		}
		spec, err = resolved.NewAggregateQuerySpec(
			catalogs, reference, storage, stableID, mustRequestPath(""), aggregate,
		)
	}
	if err == nil {
		return spec, nil
	}
	if errors.Is(err, resolved.ErrIdentityTooLarge) {
		collector.budget("", restqueryv1.IssueQuerySizeExceeded)
		return resolved.QuerySpec{}, collector.err()
	}
	return resolved.QuerySpec{}, fmt.Errorf("REST Query resolution contract failed")
}

func catalogRootParts(
	catalog api.FieldCatalog,
) (api.CatalogReference, api.ResourceStorage, api.FieldDefinition, error) {
	reference, err := catalog.Reference()
	if err != nil {
		return api.CatalogReference{}, api.ResourceStorage{}, api.FieldDefinition{}, err
	}
	storage, err := catalog.Storage()
	if err != nil {
		return api.CatalogReference{}, api.ResourceStorage{}, api.FieldDefinition{}, err
	}
	stableID, err := catalog.StableID()
	if err != nil {
		return api.CatalogReference{}, api.ResourceStorage{}, api.FieldDefinition{}, err
	}
	stableField, found, err := catalog.Field(stableID)
	if err != nil || !found {
		return api.CatalogReference{}, api.ResourceStorage{}, api.FieldDefinition{},
			fmt.Errorf("REST Query catalog stable field is invalid")
	}
	return reference, storage, stableField, nil
}
