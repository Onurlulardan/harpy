package resolved

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

type authorizedResourceKey struct {
	catalog api.CatalogReference
	path    string
}

func queryAuthorizedResources(
	spec QuerySpec,
) (map[authorizedResourceKey]api.FieldCatalog, bool) {
	if spec.Validate() != nil {
		return nil, false
	}
	catalogs, err := spec.Catalogs()
	root, rootErr := catalogs.RootCatalog()
	rootPath, pathErr := spec.RootPath()
	if err != nil || rootErr != nil || pathErr != nil {
		return nil, false
	}
	resources := make(map[authorizedResourceKey]api.FieldCatalog)
	if !addAuthorizedResource(resources, root, rootPath) {
		return nil, false
	}
	record, found, err := spec.Record()
	if err != nil {
		return nil, false
	}
	if !found {
		return resources, len(resources) == 1
	}
	includes, err := record.Includes()
	if err != nil || !addAuthorizedIncludes(resources, catalogs, includes) {
		return nil, false
	}
	return resources, len(resources) >= 1 && len(resources) <= 4
}

func addAuthorizedIncludes(
	resources map[authorizedResourceKey]api.FieldCatalog,
	catalogs api.CatalogSet,
	includes []Include,
) bool {
	for _, include := range includes {
		reference, referenceErr := include.Target()
		path, pathErr := include.Path()
		catalog, found, catalogErr := catalogs.Catalog(reference)
		if referenceErr != nil || pathErr != nil || catalogErr != nil || !found ||
			!addAuthorizedResource(resources, catalog, path) {
			return false
		}
		children, err := include.Children()
		if err != nil || !addAuthorizedIncludes(resources, catalogs, children) {
			return false
		}
	}
	return true
}

func addAuthorizedResource(
	resources map[authorizedResourceKey]api.FieldCatalog,
	catalog api.FieldCatalog,
	path RequestPath,
) bool {
	reference, referenceErr := catalog.Reference()
	value, pathErr := path.Value()
	if referenceErr != nil || pathErr != nil {
		return false
	}
	key := authorizedResourceKey{catalog: reference, path: value}
	if _, duplicate := resources[key]; duplicate {
		return false
	}
	resources[key] = catalog
	return true
}
