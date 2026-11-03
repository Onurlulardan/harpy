package baseline

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed sql/20261103000005_platform_tenant_executor_claim_clock_baseline.fresh.sql
var platformTenantExecutorClaimClockScript string

const (
	platformTenantExecutorClaimClockVersion = int64(20261103000005)
	platformTenantExecutorClaimClockFile    = "20261103000005_platform_tenant_executor_claim_clock_baseline.fresh.sql"
)

func applyPlatformTenantExecutorClaimClock(ctx context.Context, dsn string) (Result, error) {
	result, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformTenantExecutorClaimClockFile, platformTenantExecutorClaimClockScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform tenant executor claim clock: %w", err)
	}
	return result, nil
}
