package iam

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	systemtransferpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/systemtransfer"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type SystemTransferImportJobRuntime struct {
	processor jobexecution.Processor
	sealer    *DeliverySealer
	close     sync.Once
	closed    atomic.Bool
	closeErr  error
}

func OpenSystemTransferImportJobRuntime(ctx context.Context,
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	release platformasync.ReleaseCatalogs,
	catalogs platformasync.WriteCatalogs,
	cursorKeys config.CursorKeyring,
	settings config.IAMWorker,
	source systemtransfer.ExactImportSourceOpener,
) (*SystemTransferImportJobRuntime, error) {
	if ctx == nil || ctx.Err() != nil || nilIdentityDependency(pool) ||
		coordinator == nil || settings.Validate() != nil ||
		nilIdentityDependency(source) {
		return nil, systemtransfer.ErrDependencies
	}
	jobCatalog, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read IAM system import job catalog: %w", err)
	}
	policy, err := jobCatalog.Resolve(SystemImportJobKind, 1)
	if err != nil || ValidateSystemImportJobPolicy(policy) != nil {
		return nil, ErrSystemTransferJobPolicy
	}
	attempts, err := systemtransferpostgres.NewAttemptStore(coordinator)
	if err != nil {
		return nil, err
	}
	persistence, err := systemtransferpostgres.NewWorkerStore()
	if err != nil {
		return nil, err
	}
	scopes, err := newSystemTransferBatchScopes(pool, coordinator)
	if err != nil {
		return nil, err
	}
	authorization, err := NewIAMWorkerAuthorizationRuntime()
	if err != nil {
		return nil, err
	}
	authorizer, err := newSystemTransferBatchAuthorizer(authorization)
	if err != nil {
		return nil, err
	}
	receipts, err := newSystemTransferReceiptStore(scopes, authorizer, persistence)
	if err != nil {
		return nil, err
	}
	mutator, err := newSystemTransferResourceMutator(
		pool, coordinator, cursorKeys, catalogs,
	)
	if err != nil {
		return nil, err
	}
	sealer, err := OpenWorkerDeliverySealer(ctx, settings)
	if err != nil {
		return nil, err
	}
	runtime := &SystemTransferImportJobRuntime{sealer: sealer}
	invitation, err := newSystemTransferImportInvitation(
		settings, catalogs, sealer, persistence,
	)
	if err != nil {
		_ = runtime.Close()
		return nil, err
	}
	worker, err := systemtransfer.NewWorker(
		scopes, authorizer, persistence, mutator, invitation, receipts,
	)
	if err != nil {
		_ = runtime.Close()
		return nil, err
	}
	imports, err := systemtransfer.NewSystemImportRuntime(attempts, source, worker)
	if err != nil {
		_ = runtime.Close()
		return nil, err
	}
	terminal, err := newSystemTransferImportJobTerminal(
		scopes, authorizer, persistence, catalogs,
	)
	if err != nil {
		_ = runtime.Close()
		return nil, err
	}
	runtime.processor, err = newSystemTransferImportJobProcessor(
		policy, imports, terminal,
	)
	if err != nil {
		_ = runtime.Close()
		return nil, err
	}
	return runtime, nil
}

func (runtime *SystemTransferImportJobRuntime) Processor() (
	jobexecution.Processor, error,
) {
	if runtime == nil || runtime.closed.Load() || runtime.processor == nil {
		return nil, systemtransfer.ErrDependencies
	}
	return runtime.processor, nil
}

func (runtime *SystemTransferImportJobRuntime) Close() error {
	if runtime == nil {
		return nil
	}
	runtime.close.Do(func() {
		runtime.closed.Store(true)
		if runtime.sealer == nil {
			runtime.closeErr = systemtransfer.ErrDependencies
			return
		}
		runtime.closeErr = errors.Join(runtime.closeErr, runtime.sealer.Close())
	})
	return runtime.closeErr
}
