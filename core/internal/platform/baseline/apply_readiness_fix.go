package baseline

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed sql/20261103000001_platform_tenant_executor_baseline_readiness_fix.fresh.sql
var platformTenantExecutorReadinessFixScript string

//go:embed sql/20261103000002_platform_tenant_executor_baseline_claim.fresh.sql
var platformTenantExecutorClaimFixScript string

const (
	platformTenantExecutorReadinessFixVersion = int64(20261103000001)
	platformTenantExecutorReadinessFixFile    = "20261103000001_platform_tenant_executor_baseline_readiness_fix.fresh.sql"
	platformTenantExecutorClaimFixVersion     = int64(20261103000002)
	platformTenantExecutorClaimFixFile        = "20261103000002_platform_tenant_executor_baseline_claim.fresh.sql"
)

func applyPlatformTenantExecutorReadinessFix(ctx context.Context, dsn string) (Result, error) {
	readiness, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformTenantExecutorReadinessFixFile, platformTenantExecutorReadinessFixScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform tenant executor readiness fix: %w", err)
	}
	claim, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformTenantExecutorClaimFixFile, platformTenantExecutorClaimFixScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform tenant executor claim fix: %w", err)
	}
	return combineBaselineResults(readiness, claim), nil
}
