package worker

import (
	"context"
	"fmt"

	iam "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type iamExportReconciliationRecoveryStore struct {
	asyncpostgres.WorkerStore
	recovery iam.SystemTransferExportReconciliationRecovery
}

func newIAMExportReconciliationRecoveryStore(
	delegate asyncpostgres.WorkerStore,
	recovery iam.SystemTransferExportReconciliationRecovery,
) (*iamExportReconciliationRecoveryStore, error) {
	if delegate == nil || recovery == nil {
		return nil, fmt.Errorf("IAM export reconciliation recovery is incomplete")
	}
	return &iamExportReconciliationRecoveryStore{
		WorkerStore: delegate, recovery: recovery,
	}, nil
}

func (store *iamExportReconciliationRecoveryStore) RecoverExpiredJobs(
	ctx context.Context,
	authority asyncpostgres.JobAuthority,
	limit int,
) (asyncpostgres.JobRecoveryResult, error) {
	tenantID, tenantScoped := authority.TenantID()
	if tenantScoped {
		if err := store.recovery.RecoverExpiredSystemTransferExports(
			ctx, tenantID, limit,
		); err != nil {
			return asyncpostgres.JobRecoveryResult{}, fmt.Errorf(
				"recover IAM export reconciliations: %w", err)
		}
	}
	return store.WorkerStore.RecoverExpiredJobs(ctx, authority, limit)
}
