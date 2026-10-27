package systemresourcequery

import (
	"context"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

// RunValue executes a validated owner request against one exact IAM catalog
// snapshot. Transport adapters are deliberately outside this API.
func (service *Service) RunValue(
	ctx context.Context,
	scope transaction.TenantScope,
	published catalog.Published,
	request publicresource.QueryRequest,
) (Result, error) {
	value, validRequestType := request.(restquery.QueryRequest)
	set, catalogErr := published.CatalogSet()
	if service == nil || service.engine == nil || ctx == nil ||
		request == nil || request.Validate() != nil || !validRequestType || catalogErr != nil {
		return Result{}, ErrContract
	}
	result, err := service.engine.ExecuteCatalog(ctx, scope, set, value)
	if err != nil {
		return Result{}, NormalizeError(err)
	}
	return newResult(result, ""), nil
}
