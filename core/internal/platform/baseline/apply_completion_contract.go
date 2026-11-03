package baseline

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed sql/20261103000006_platform_tenant_executor_completion_contract_baseline.fresh.sql
var platformTenantExecutorCompletionContractScript string

const (
	platformTenantExecutorCompletionContractVersion = int64(20261103000006)
	platformTenantExecutorCompletionContractFile    = "20261103000006_platform_tenant_executor_completion_contract_baseline.fresh.sql"
)

func applyPlatformTenantExecutorCompletionContract(ctx context.Context, dsn string) (Result, error) {
	result, err := apply(
		ctx, dsn, "platform, pg_catalog",
		platformTenantExecutorCompletionContractFile,
		platformTenantExecutorCompletionContractScript,
	)
	if err != nil {
		return Result{}, fmt.Errorf("apply platform tenant executor completion contract: %w", err)
	}
	return result, nil
}
