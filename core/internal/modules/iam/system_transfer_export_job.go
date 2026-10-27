package iam

import (
	"context"
	"errors"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type systemTransferExportJobProcessor struct {
	policy     jobsv1.KindPolicy
	attempts   systemtransfer.ExportAttemptLoader
	scopes     systemtransfer.FreshExportScopes
	authorizer systemtransfer.ExportCheckpointAuthorizer
	stores     systemtransfer.ExportWorkerStoreBinder
	snapshots  systemtransfer.ExportSnapshotProvider
	spools     systemtransfer.ExportPlaintextSpoolAllocator
	artifacts  systemtransfer.ExportArtifactSealer
	jobStore   asyncpostgres.WorkerStore
	preEffect  asyncpostgres.ExternalPreEffectStore
}

func newSystemTransferExportJobProcessor(
	policy jobsv1.KindPolicy,
	attempts systemtransfer.ExportAttemptLoader,
	scopes systemtransfer.FreshExportScopes,
	authorizer systemtransfer.ExportCheckpointAuthorizer,
	stores systemtransfer.ExportWorkerStoreBinder,
	snapshots systemtransfer.ExportSnapshotProvider,
	spools systemtransfer.ExportPlaintextSpoolAllocator,
	artifacts systemtransfer.ExportArtifactSealer,
	jobStore asyncpostgres.WorkerStore,
) (*systemTransferExportJobProcessor, error) {
	preEffect, err := asyncpostgres.BindExternalPreEffect(jobStore)
	dependencies := []any{attempts, scopes, authorizer, stores, snapshots, spools, artifacts}
	if ValidateSystemExportJobPolicy(policy) != nil || err != nil || jobStore == nil {
		return nil, ErrSystemTransferJobPolicy
	}
	for _, dependency := range dependencies {
		if nilIdentityDependency(dependency) {
			return nil, ErrSystemTransferJobPolicy
		}
	}
	return &systemTransferExportJobProcessor{policy: policy, attempts: attempts,
		scopes: scopes, authorizer: authorizer, stores: stores, snapshots: snapshots,
		spools: spools, artifacts: artifacts, jobStore: jobStore, preEffect: preEffect}, nil
}

func (processor *systemTransferExportJobProcessor) Policy() jobsv1.KindPolicy {
	if processor == nil {
		return jobsv1.KindPolicy{}
	}
	return processor.policy
}

func (processor *systemTransferExportJobProcessor) Process(
	ctx context.Context,
	raw asyncpostgres.JobClaim,
) error {
	if processor == nil || ctx == nil || processor.jobStore == nil ||
		processor.preEffect == nil {
		return ErrSystemTransferJobPolicy
	}
	claim, err := decodeSystemTransferExportClaim(raw, processor.policy)
	if err != nil {
		return err
	}
	lease := &systemTransferExportJobLease{
		store: processor.jobStore, raw: raw, decoded: claim,
	}
	values := systemTransferExportValues{}
	worker, err := systemtransfer.NewSystemExportWorker(
		values, processor.attempts, lease, processor.scopes, processor.authorizer,
		processor.stores, processor.snapshots, values, processor.spools, processor.artifacts,
	)
	if err != nil {
		return fmt.Errorf("build IAM system export worker: %w", err)
	}
	result, executeErr := worker.ExecuteExport(ctx, claim)
	if result.State == systemtransfer.ExportExecutionLeaseLost ||
		errors.Is(executeErr, systemtransfer.ErrExportLeaseLost) ||
		ctx.Err() != nil && result.State == systemtransfer.ExportExecutionRetry {
		return asyncpostgres.ErrJobLeaseLost
	}
	switch result.State {
	case systemtransfer.ExportExecutionReady,
		systemtransfer.ExportExecutionFailed,
		systemtransfer.ExportExecutionCancelled:
		if executeErr == nil {
			return nil
		}
		return processor.persistPreEffectResult(ctx, raw, result, executeErr)
	case systemtransfer.ExportExecutionRetry:
		if executeErr == nil {
			return ErrSystemTransferJobPolicy
		}
		return processor.persistPreEffectResult(ctx, raw, result, executeErr)
	case systemtransfer.ExportExecutionOutcomeUnknown:
		if executeErr != nil {
			return fmt.Errorf("persist IAM export unknown outcome: %w", executeErr)
		}
		return nil
	default:
		return fmt.Errorf("execute IAM system export: %w", executeErr)
	}
}
