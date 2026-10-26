package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type Include struct {
	relation        api.RelationDefinition
	target          api.CatalogReference
	targetStorage   api.ResourceStorage
	relationStorage api.RelationStorage
	stableID        api.FieldDefinition
	path            RequestPath
	projections     []Projection
	where           *Filter
	sorts           []Sort
	limit           int
	children        []Include
	validated       bool
}

func NewInclude(
	relation api.RelationDefinition,
	target api.CatalogReference,
	targetStorage api.ResourceStorage,
	relationStorage api.RelationStorage,
	stableID api.FieldDefinition,
	path RequestPath,
	projections []Projection,
	where *Filter,
	sorts []Sort,
	limit int,
	children []Include,
) (Include, error) {
	include := Include{
		relation: relation, target: target, targetStorage: targetStorage,
		relationStorage: relationStorage, stableID: stableID, path: path,
		projections: append([]Projection{}, projections...), limit: limit,
		sorts: append([]Sort{}, sorts...), children: cloneIncludes(children), validated: true,
	}
	if where != nil {
		owned := where.clone()
		include.where = &owned
	}
	if err := include.Validate(); err != nil {
		return Include{}, err
	}
	return include, nil
}

func (include Include) Validate() error {
	if !include.validated || include.relation.Validate() != nil ||
		include.target.Validate() != nil || include.targetStorage.Validate() != nil ||
		include.relationStorage.Validate() != nil ||
		include.stableID.Validate() != nil || include.path.Validate() != nil ||
		include.projections == nil || include.sorts == nil ||
		include.children == nil || !validIncludeLimit(include) {
		return fmt.Errorf("resolved REST Query include is invalid")
	}
	if include.where != nil && include.where.Validate() != nil {
		return fmt.Errorf("resolved REST Query include is invalid")
	}
	if !validIncludeSorts(include) {
		return fmt.Errorf("resolved REST Query include is invalid")
	}
	target, err := include.relation.Target()
	relationStorage, storageErr := include.relation.Storage()
	if err != nil || !sameCatalogReference(target, include.target) ||
		storageErr != nil || relationStorage != include.relationStorage ||
		!validIncludeProjections(include.projections, include.stableID) {
		return fmt.Errorf("resolved REST Query include is invalid")
	}
	for _, child := range include.children {
		if child.Validate() != nil {
			return fmt.Errorf("resolved REST Query include is invalid")
		}
	}
	return nil
}

func validIncludeLimit(include Include) bool {
	enabled, err := include.relation.IncludeEnabled()
	maximum, maximumErr := include.relation.MaximumIncludedRecords()
	cardinality, cardinalityErr := include.relation.Cardinality()
	if err != nil || maximumErr != nil || cardinalityErr != nil || !enabled ||
		include.limit < 1 || include.limit > maximum {
		return false
	}
	return cardinality != api.RelationCardinalityToOne || include.limit == 1
}
