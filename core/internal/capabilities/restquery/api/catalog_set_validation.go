package api

import "github.com/google/uuid"

type catalogIdentityKey struct {
	owner              string
	resourceID         uuid.UUID
	publishedVersionID uuid.UUID
}

type exactCatalogKey struct {
	identity catalogIdentityKey
	hash     CatalogHash
}

func catalogSetIndex(
	catalogs []FieldCatalog,
) (map[exactCatalogKey]struct{}, bool) {
	exact := make(map[exactCatalogKey]struct{}, len(catalogs))
	identities := make(map[catalogIdentityKey]struct{}, len(catalogs))
	for _, catalog := range catalogs {
		if catalog.Validate() != nil {
			return nil, false
		}
		key := exactCatalogKeyOf(catalog.reference)
		if _, duplicate := identities[key.identity]; duplicate {
			return nil, false
		}
		identities[key.identity] = struct{}{}
		exact[key] = struct{}{}
	}
	return exact, catalogStorageMappingsValid(catalogs)
}

func catalogRelationsResolve(
	catalogs []FieldCatalog,
	exact map[exactCatalogKey]struct{},
) bool {
	for _, catalog := range catalogs {
		for _, relation := range catalog.relations {
			if relation.Validate() != nil {
				return false
			}
			if _, found := exact[exactCatalogKeyOf(relation.target)]; !found {
				return false
			}
		}
	}
	return true
}

func exactCatalogKeyOf(reference CatalogReference) exactCatalogKey {
	return exactCatalogKey{
		identity: catalogIdentityKey{
			owner: reference.owner, resourceID: reference.resourceID,
			publishedVersionID: reference.publishedVersionID,
		},
		hash: reference.catalogHash,
	}
}
