package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func (compiler Compiler) extractRecord(
	plan resolved.AuthorizedPlan,
	expectedCount api.CountMode,
) (recordInput, error) {
	if expectedCount == api.CountModeExact {
		authorized, err := plan.ExactCountAuthorized()
		if err != nil || !authorized {
			return recordInput{}, ErrContract
		}
	}
	spec, err := plan.Spec()
	if err != nil {
		return recordInput{}, ErrContract
	}
	mode, err := spec.Mode()
	if err != nil || mode != api.QueryModeRecord {
		return recordInput{}, ErrContract
	}
	query, found, err := spec.Record()
	if err != nil || !found ||
		validateRecordCountShape(query, expectedCount) != nil {
		return recordInput{}, ErrContract
	}
	storage, err := spec.Storage()
	if err != nil {
		return recordInput{}, ErrContract
	}
	reference, referenceErr := spec.Reference()
	rootPath, pathErr := spec.RootPath()
	rootPathValue, pathValueErr := rootPath.Value()
	identity, identityErr := spec.IdentityHash()
	if referenceErr != nil || pathErr != nil || pathValueErr != nil || identityErr != nil {
		return recordInput{}, ErrContract
	}
	visibility, err := rootVisibility(plan, reference, rootPathValue)
	if err != nil {
		return recordInput{}, err
	}
	binding, err := newCursorBinding(plan, reference, identity)
	if err != nil {
		return recordInput{}, err
	}
	continuation, err := continuationIdentity(query)
	if err != nil {
		return recordInput{}, err
	}
	provenance, err := newStatementBinding(binding, continuation)
	if err != nil {
		return recordInput{}, err
	}
	context, err := cursorContext(binding, query)
	if err != nil {
		return recordInput{}, err
	}
	last, err := compiler.decodeCursor(context, query)
	if err != nil {
		return recordInput{}, err
	}
	input := recordInput{
		storage: storage, query: query, visibility: visibility, last: last,
		context: context, binding: provenance,
	}
	if input.validate() != nil {
		return recordInput{}, ErrContract
	}
	return input, nil
}

func rootVisibility(
	plan resolved.AuthorizedPlan,
	reference api.CatalogReference,
	pathValue string,
) (resolved.ResourceVisibility, error) {
	visibilities, visibilityErr := plan.Visibilities()
	if reference.Validate() != nil || visibilityErr != nil {
		return resolved.ResourceVisibility{}, ErrContract
	}
	var result resolved.ResourceVisibility
	found := false
	for _, visibility := range visibilities {
		catalog, catalogErr := visibility.Catalog()
		visibilityPath, visibilityPathErr := visibility.Path()
		visibilityValue, visibilityValueErr := visibilityPath.Value()
		if catalogErr != nil || visibilityPathErr != nil || visibilityValueErr != nil {
			return resolved.ResourceVisibility{}, ErrContract
		}
		if catalog == reference && visibilityValue == pathValue {
			if found {
				return resolved.ResourceVisibility{}, ErrContract
			}
			result, found = visibility, true
		}
	}
	if !found {
		return resolved.ResourceVisibility{}, ErrContract
	}
	return result, nil
}
