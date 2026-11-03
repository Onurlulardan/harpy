package baseline

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed sql/20261103000003_platform_tenant_executor_baseline.fresh.sql
var platformTenantExecutorBaselineScript string

const (
	platformTenantExecutorBaselineVersion = int64(20261103000003)
	platformTenantExecutorBaselineFile    = "20261103000003_platform_tenant_executor_baseline.fresh.sql"
)

func applyPlatformTenantExecutorBaseline(ctx context.Context, dsn string) (Result, error) {
	result, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformTenantExecutorBaselineFile, platformTenantExecutorBaselineScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform tenant executor baseline: %w", err)
	}
	return result, nil
}
