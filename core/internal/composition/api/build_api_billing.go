package api

import (
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildAPIBilling(
	settings config.API, pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs, iamRuntime iamAPIRuntime, report httpedge.ErrorSink,
) ([][]httpedge.Route, error) {
	billingHTTP, err := buildBillingHTTP(
		settings, pool, catalogs, iamRuntime.billingAccess, report,
	)
	if err != nil {
		return nil, err
	}
	return buildBillingRouteGroups(billingHTTP)
}
