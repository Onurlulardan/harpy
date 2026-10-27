package billing

import (
	catalogpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/catalogquery"
	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func BindCatalogQueryStore(scope platformtx.GlobalScope) (catalogquery.Store, error) {
	return catalogpostgres.Bind(scope)
}
