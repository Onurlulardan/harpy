package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *SystemTransferRuntime) enqueueSystemTransfer(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource SystemTransferResource,
	transferID, jobID uuid.UUID,
	kind string,
	now time.Time,
) error {
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return ErrSystemTransferUnavailable
	}
	jobs, _, err := runtime.catalogs.Snapshot()
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	policy, err := jobs.Resolve(kind, 1)
	if err != nil || !systemTransferPolicyValid(kind, policy) {
		return ErrSystemTransferUnavailable
	}
	publicResource, err := systemTransferResource(resource)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	reference, err := systemtransfer.JobReference(publicResource, transferID)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	digest, err := systemtransfer.JobReferenceSHA256(reference)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	tenantID := scope.TenantID().String()
	value := jobsv1.Enqueue{
		JobID: jobID.String(), Kind: kind, PayloadRevision: 1, TenantID: &tenantID,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenantID,
			Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
			WriterOrigin: jobsv1.WriterV2, GroupID: authority.GroupID(),
			GroupGeneration: authority.GroupGeneration(),
		},
		PayloadReference: reference, PayloadSHA256: digest,
		RunAt: now.Format(time.RFC3339Nano), Priority: 0,
		DedupeKey:       "iam-system-transfer:" + kind + ":" + transferID.String(),
		MaximumAttempts: policy.MaximumAttempts,
	}
	if kind == SystemExportJobKind {
		effect := "iam-system-export:" + transferID.String()
		value.EffectIdempotencyKey = &effect
	}
	job, err := jobsv1.NewEnqueue(value, jobs)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	writer, err := asyncpostgres.Bind(scope, authority, runtime.catalogs)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	result, err := writer.EnqueueJob(ctx, job)
	if err != nil || result.ID != jobID.String() {
		return ErrSystemTransferUnavailable
	}
	return nil
}

func (runtime *SystemTransferRuntime) cancelSystemTransferJob(
	ctx context.Context,
	scope platformtx.TenantScope,
	jobID uuid.UUID,
	now time.Time,
	code string,
) error {
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return ErrSystemTransferUnavailable
	}
	digest := sha256.Sum256([]byte(code))
	request, err := asyncpostgres.NewOwnedJobCancellation(
		jobID, now, code, hex.EncodeToString(digest[:]),
	)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	writer, err := asyncpostgres.Bind(scope, authority, runtime.catalogs)
	if err != nil {
		return ErrSystemTransferUnavailable
	}
	_, err = writer.CancelOwnedJob(ctx, request)
	return systemTransferError(err)
}

func systemTransferPolicyValid(kind string, policy jobsv1.KindPolicy) bool {
	if kind == SystemExportJobKind {
		return ValidateSystemExportJobPolicy(policy) == nil
	}
	return kind == SystemImportJobKind && ValidateSystemImportJobPolicy(policy) == nil
}
