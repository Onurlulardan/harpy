package storagecatalog

import workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"

func New() (workquery.CatalogPort, error) {
	catalogs := make([]workquery.ResolvedCatalog, 0, len(allQuerySpecs()))
	for _, spec := range allQuerySpecs() {
		catalog, err := newCatalog(spec)
		if err != nil {
			return nil, err
		}
		catalogs = append(catalogs, catalog)
	}
	return func() ([]workquery.ResolvedCatalog, error) {
		return append([]workquery.ResolvedCatalog{}, catalogs...), nil
	}, nil
}
