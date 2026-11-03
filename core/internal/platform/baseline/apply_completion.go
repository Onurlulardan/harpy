package baseline

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed sql/20261103000004_platform_tenant_executor_completion_baseline.fresh.sql
var platformTenantExecutorCompletionScript string

const (
	platformTenantExecutorCompletionVersion = int64(20261103000004)
	platformTenantExecutorCompletionFile    = "20261103000004_platform_tenant_executor_completion_baseline.fresh.sql"
)

func applyPlatformTenantExecutorCompletion(ctx context.Context, dsn string) (Result, error) {
	result, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformTenantExecutorCompletionFile, platformTenantExecutorCompletionScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform tenant executor completion: %w", err)
	}
	return result, nil
}
