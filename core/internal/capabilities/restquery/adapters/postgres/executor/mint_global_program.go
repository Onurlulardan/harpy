package executor

import (
	"context"

	queryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func mintGlobalProgram(
	ctx context.Context,
	scope transaction.GlobalScope,
	reference queryapi.CatalogReference,
	run func(context.Context, execution.GlobalSnapshotDriver) error,
) (execution.GlobalProgram, error) {
	return execution.MintGlobalProgram(ctx, scope, reference, run)
}
