package resolved

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

func specRootMatchesCatalogs(spec QuerySpec) bool {
	rootReference, err := spec.catalogs.Root()
	if err != nil || !sameCatalogReference(rootReference, spec.reference) {
		return false
	}
	root, err := spec.catalogs.RootCatalog()
	if err != nil {
		return false
	}
	storage, storageErr := root.Storage()
	stableID, stableErr := root.StableID()
	resolvedStableID, resolvedStableErr := spec.stableID.ID()
	return storageErr == nil && stableErr == nil && resolvedStableErr == nil &&
		storage == spec.storage && stableID == resolvedStableID
}

func validStableField(field api.FieldDefinition) bool {
	fieldType, err := field.Type()
	if err != nil {
		return false
	}
	kind, err := fieldType.Kind()
	return err == nil && kind == api.LogicalTypeUUID
}

func validRecordSpec(
	record RecordQuery, root api.CatalogReference, stableID api.FieldDefinition,
) bool {
	stable, valid := fieldIdentifier(stableID)
	if !valid || !recordStableProjection(record.projections, stable) ||
		!validStableFinalSort(record.sorts, stableID, implicitSortCount(record.sorts)) {
		return false
	}
	projectionCount, edges, valid := validateIncludeTree(
		record.includes, root, map[resourceNodeKey]struct{}{}, 1,
	)
	if !valid || visibleProjectionCount(record.projections)+projectionCount > 50 || edges > 3 {
		return false
	}
	return validRecordFilterBudget(record)
}

func recordStableProjection(projections []Projection, stable uuid.UUID) bool {
	stableCount, hiddenCount := 0, 0
	for _, projection := range projections {
		identifier, valid := fieldIdentifier(projection.field)
		if !valid {
			return false
		}
		if projection.hidden {
			hiddenCount++
			if identifier != stable {
				return false
			}
		}
		if identifier == stable {
			stableCount++
		}
	}
	return stableCount == 1 && hiddenCount <= 1
}

func implicitSortCount(sorts []Sort) int {
	count := 0
	for _, sort := range sorts {
		if sort.implicit {
			count++
		}
	}
	return count
}

func visibleProjectionCount(projections []Projection) int {
	count := 0
	for _, projection := range projections {
		if !projection.hidden {
			count++
		}
	}
	return count
}

type resourceNodeKey struct {
	owner      string
	resourceID uuid.UUID
}
