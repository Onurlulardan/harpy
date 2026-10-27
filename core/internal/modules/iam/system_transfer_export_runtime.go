package iam

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	systemtransferpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/systemtransfer"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type SystemTransferExportJobRuntime struct {
	processors []jobexecution.Processor
	close      sync.Once
	closed     atomic.Bool
}

func OpenSystemTransferExportJobRuntime(
	ctx context.Context,
	pool IdentityHTTPPlatform,
	coordinator *platformtx.Coordinator,
	release platformasync.ReleaseCatalogs,
	cursorKeys config.CursorKeyring,
	jobStore asyncpostgres.WorkerStore,
	artifacts SystemTransferArtifacts,
) (*SystemTransferExportJobRuntime, error) {
	if ctx == nil || ctx.Err() != nil || nilIdentityDependency(pool) ||
		coordinator == nil || nilIdentityDependency(release) || jobStore == nil ||
		nilIdentityDependency(artifacts) {
		return nil, systemtransfer.ErrExportUnavailable
	}
	jobs, _, err := release.Snapshot()
	if err != nil {
		return nil, fmt.Errorf("read IAM export job catalog: %w", err)
	}
	policy, err := jobs.Resolve(SystemExportJobKind, 1)
	if err != nil || ValidateSystemExportJobPolicy(policy) != nil {
		return nil, ErrSystemTransferJobPolicy
	}
	attempts, err := systemtransferpostgres.NewOutboundAttemptStore(coordinator)
	if err != nil {
		return nil, err
	}
	stores, err := systemtransferpostgres.NewOutboundStore()
	if err != nil {
		return nil, err
	}
	reconciliation, err := systemtransferpostgres.NewReconciliationStore(coordinator)
	if err != nil {
		return nil, err
	}
	scopes, err := newSystemTransferExportScopes(pool, coordinator)
	if err != nil {
		return nil, err
	}
	authorization, err := NewIAMWorkerAuthorizationRuntime()
	if err != nil {
		return nil, err
	}
	authorizer, err := newSystemTransferExportAuthorizer(authorization)
	if err != nil {
		return nil, err
	}
	keys, err := identityQueryKeysFromConfig(cursorKeys)
	if err != nil {
		return nil, err
	}
	snapshots, err := newSystemTransferExportSnapshots(scopes, authorizer, keys)
	if err != nil {
		return nil, err
	}
	exports, err := newSystemTransferExportJobProcessor(
		policy, attempts, scopes, authorizer, stores, snapshots,
		artifacts, artifacts, jobStore,
	)
	if err != nil {
		return nil, err
	}
	reconcile, err := NewSystemTransferExportReconciliationProcessor(
		release, reconciliation, artifacts,
	)
	if err != nil {
		return nil, err
	}
	return &SystemTransferExportJobRuntime{
		processors: []jobexecution.Processor{exports, reconcile},
	}, nil
}

func (runtime *SystemTransferExportJobRuntime) Processors() (
	[]jobexecution.Processor,
	error,
) {
	if runtime == nil || runtime.closed.Load() || len(runtime.processors) != 2 {
		return nil, systemtransfer.ErrExportUnavailable
	}
	return append([]jobexecution.Processor(nil), runtime.processors...), nil
}

func (runtime *SystemTransferExportJobRuntime) Close() error {
	if runtime == nil {
		return nil
	}
	runtime.close.Do(func() {
		runtime.closed.Store(true)
	})
	return nil
}
