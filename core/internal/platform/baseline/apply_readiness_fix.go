package baseline

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed sql/20261103000001_platform_tenant_executor_baseline_readiness_fix.fresh.sql
var platformTenantExecutorReadinessFixScript string

const (
	platformTenantExecutorReadinessFixVersion = int64(20261103000001)
	platformTenantExecutorReadinessFixFile    = "20261103000001_platform_tenant_executor_baseline_readiness_fix.fresh.sql"
)

func applyPlatformTenantExecutorReadinessFix(ctx context.Context, dsn string) (Result, error) {
	result, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformTenantExecutorReadinessFixFile, platformTenantExecutorReadinessFixScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform tenant executor readiness fix: %w", err)
	}
	return result, nil
}
