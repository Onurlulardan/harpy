package work

import (
	"context"

	workhttp "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/http"
	workpostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *PlatformHTTPRuntime) RecordCompatibilityHit(ctx context.Context,
	hit HTTPCompatibilityHit,
) error {
	if omitCompatibilityHit(hit) {
		return nil
	}
	operation, valid := compatibilityOperation(hit)
	contract, found := workhttp.ExpectedAccess(operation)
	if !valid || !found || !contract.Compatibility ||
		!hit.Access.ValidFor(contract.Authority, operation, true) {
		return ErrHTTPForbidden
	}
	values, err := runtime.resolve(ctx, hit.Access)
	if err != nil {
		return err
	}
	return runtime.command(ctx, hit.Access, values, "work.compatibility.hit",
		func(operation context.Context, scope platformtx.TenantScope) error {
			return runtime.hits.Record(operation, workpostgres.BridgeScope(scope),
				hit.RouteContract, hit.ConsumerID, hit.BuildID)
		})
}

func omitCompatibilityHit(hit HTTPCompatibilityHit) bool {
	return hit.ConsumerID == "" || hit.BuildID == ""
}
