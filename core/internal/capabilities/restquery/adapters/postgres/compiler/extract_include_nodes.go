package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func collectIncludeNodeInputs(
	plan resolved.AuthorizedPlan,
) ([]includeNodeInput, includePlanBinding, error) {
	binding, err := newIncludePlanBinding(plan)
	if err != nil {
		return nil, includePlanBinding{}, ErrContract
	}
	spec, err := plan.Spec()
	if err != nil {
		return nil, includePlanBinding{}, ErrContract
	}
	mode, modeErr := spec.Mode()
	query, found, queryErr := spec.Record()
	rootCatalog, catalogErr := spec.Reference()
	rootStable, stableErr := spec.StableID()
	if modeErr != nil || mode != api.QueryModeRecord || queryErr != nil || !found ||
		catalogErr != nil || stableErr != nil {
		return nil, includePlanBinding{}, ErrContract
	}
	includes, err := query.Includes()
	if err != nil || len(includes) < 1 || len(includes) > 3 {
		return nil, includePlanBinding{}, ErrContract
	}
	queue := make([]includeNodeInput, 0, 3)
	appendIncludeInputs(
		&queue, includes, rootCatalog, rootStable, 1, -1,
	)
	for index := 0; index < len(queue); index++ {
		if len(queue) > 3 {
			return nil, includePlanBinding{}, ErrContract
		}
		current := queue[index]
		children, err := current.include.Children()
		if err != nil || current.depth == 2 && len(children) != 0 {
			return nil, includePlanBinding{}, ErrContract
		}
		if len(children) == 0 {
			continue
		}
		target, targetErr := current.include.Target()
		stable, stableErr := current.include.StableID()
		if targetErr != nil || stableErr != nil {
			return nil, includePlanBinding{}, ErrContract
		}
		appendIncludeInputs(
			&queue, children, target, stable, current.depth+1, current.ordinal,
		)
	}
	if len(queue) < 1 || len(queue) > 3 {
		return nil, includePlanBinding{}, ErrContract
	}
	return queue, binding, nil
}

func appendIncludeInputs(
	queue *[]includeNodeInput,
	includes []resolved.Include,
	sourceCatalog api.CatalogReference,
	sourceStable api.FieldDefinition,
	depth int,
	parentOrdinal int,
) {
	for sibling, include := range includes {
		*queue = append(*queue, includeNodeInput{
			include: include, sourceCatalog: sourceCatalog, sourceStable: sourceStable,
			depth: depth, ordinal: len(*queue), parentOrdinal: parentOrdinal,
			siblingOrdinal: sibling,
		})
	}
}
