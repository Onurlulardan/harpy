package iam

import (
	"context"
	"fmt"

	systemtransferpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/systemtransfer"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type SystemTransferExportReconciliationRecovery interface {
	RecoverExpiredSystemTransferExports(context.Context, uuid.UUID, int) error
}

type systemTransferExportReconciliationRecovery struct {
	store *systemtransferpostgres.ExportReconciliationStore
}

func OpenSystemTransferExportReconciliationRecovery(
	coordinator *platformtx.Coordinator,
) (SystemTransferExportReconciliationRecovery, error) {
	store, err := systemtransferpostgres.NewReconciliationStore(coordinator)
	if err != nil {
		return nil, fmt.Errorf("open IAM export reconciliation recovery: %w", err)
	}
	return &systemTransferExportReconciliationRecovery{store: store}, nil
}

func (recovery *systemTransferExportReconciliationRecovery) RecoverExpiredSystemTransferExports(
	ctx context.Context,
	tenantID uuid.UUID,
	limit int,
) error {
	if recovery == nil || recovery.store == nil {
		return systemtransfer.ErrExportUnavailable
	}
	_, err := recovery.store.RecoverExpiredExportReconciliations(
		ctx, tenantID, limit,
	)
	return err
}
