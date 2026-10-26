package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func mintProgram(
	ctx context.Context,
	scope transaction.TenantScope,
	catalog api.CatalogReference,
	run func(context.Context, execution.SnapshotDriver) error,
) (execution.Program, error) {
	return execution.MintProgram(ctx, scope, catalog, run)
}
