package api

import (
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

type catalogStorageKind uint8

const (
	catalogResourceStorage catalogStorageKind = iota + 1
	catalogRelationStorage
)

type catalogStorageOwner struct {
	kind       catalogStorageKind
	owner      string
	resourceID uuid.UUID
	relationID uuid.UUID
}

func catalogStorageMappingsValid(catalogs []FieldCatalog) bool {
	physical := make(map[sqlidentifier.Token]catalogStorageOwner)
	for _, catalog := range catalogs {
		owner := catalogStorageOwner{
			kind: catalogResourceStorage, owner: catalog.reference.owner,
			resourceID: catalog.reference.resourceID,
		}
		if !admitCatalogStorage(physical, catalog.storage.table, owner) {
			return false
		}
		for _, relation := range catalog.relations {
			relationOwner := catalogStorageOwner{
				kind: catalogRelationStorage, owner: catalog.reference.owner,
				resourceID: catalog.reference.resourceID,
				relationID: relation.relationID,
			}
			if !admitCatalogStorage(physical, relation.storage.table, relationOwner) {
				return false
			}
		}
	}
	return true
}

func admitCatalogStorage(
	physical map[sqlidentifier.Token]catalogStorageOwner,
	table sqlidentifier.Token,
	owner catalogStorageOwner,
) bool {
	if table.Validate() != nil {
		return false
	}
	if existing, found := physical[table]; found {
		return existing == owner
	}
	physical[table] = owner
	return true
}
