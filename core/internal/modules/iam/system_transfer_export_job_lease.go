package iam

import (
	"context"
	"errors"
	"time"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

const systemTransferExportFinalizationTimeout = 5 * time.Second

type systemTransferExportJobLease struct {
	store   asyncpostgres.WorkerStore
	raw     asyncpostgres.JobClaim
	decoded systemtransfer.ExportJobClaim
}

func (lease *systemTransferExportJobLease) CheckpointExportLease(
	ctx context.Context,
	claim systemtransfer.ExportJobClaim,
) error {
	if lease == nil || lease.store == nil || claim.Validate() != nil ||
		claim != lease.decoded {
		return systemtransfer.ErrExportLeaseLost
	}
	heartbeatContext := ctx
	cancel := func() {}
	if ctx == nil || ctx.Err() != nil {
		if ctx == nil {
			ctx = context.Background()
		}
		heartbeatContext, cancel = context.WithTimeout(
			context.WithoutCancel(ctx), systemTransferExportFinalizationTimeout,
		)
	}
	defer cancel()
	current, err := lease.store.HeartbeatJob(heartbeatContext, lease.raw)
	if errors.Is(err, asyncpostgres.ErrJobLeaseLost) {
		return systemtransfer.ErrExportLeaseLost
	}
	if err != nil {
		return errors.Join(systemtransfer.ErrExportUnavailable, err)
	}
	if current.CancelRequested() {
		return systemtransfer.ErrExportCancelled
	}
	if ctx.Err() != nil {
		return systemtransfer.ErrExportLeaseLost
	}
	return nil
}

var _ systemtransfer.ExportJobLease = (*systemTransferExportJobLease)(nil)
