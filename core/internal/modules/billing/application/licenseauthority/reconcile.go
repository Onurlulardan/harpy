package licenseauthority

import (
	"context"
	"errors"
	"fmt"
)

type Reconciler struct {
	store    Store
	source   Source
	keys     KeyResolver
	versions []ComponentVersion
}

func NewLicenseReconciler(
	store Store,
	source Source,
	keys KeyResolver,
	versions []ComponentVersion,
) (*Reconciler, error) {
	copied := append([]ComponentVersion(nil), versions...)
	if store == nil || source == nil || keys == nil || len(copied) == 0 {
		return nil, fmt.Errorf("license reconciler dependencies are invalid")
	}
	return &Reconciler{
		store: store, source: source, keys: keys, versions: copied,
	}, nil
}

func (reconciler *Reconciler) Reconcile(
	ctx context.Context,
	command RefreshCommand,
) (ReconcileResult, error) {
	claim, err := reconciler.store.AcquireRefreshLease(ctx, command, reconciler.versions)
	if err != nil {
		return ReconcileResult{}, fmt.Errorf("claim license refresh: %w", err)
	}
	if !validRefreshClaim(command, claim) {
		return ReconcileResult{}, ErrAuthorityState
	}
	if !validIssueInput(claim, reconciler.versions) {
		return ReconcileResult{}, ErrIssueRequest
	}
	request := licenseIssueRequest(claim, reconciler.versions)
	document, err := reconciler.source.IssueSnapshot(ctx, request)
	if err != nil {
		return ReconcileResult{}, reconciler.persistFailure(ctx, claim, FailureSource, err)
	}
	snapshot, err := Verify(
		ctx, document, claim.Binding, claim.DatabaseNow, reconciler.keys,
	)
	if err != nil {
		return ReconcileResult{}, reconciler.persistFailure(
			ctx, claim, FailureVerification, err,
		)
	}
	decision, err := SequenceDecisionFor(claim.Head, snapshot)
	if err != nil {
		kind := FailureConflict
		if errors.Is(err, ErrSequenceRollback) {
			kind = FailureRollback
		}
		return ReconcileResult{}, reconciler.persistFailure(ctx, claim, kind, err)
	}
	if err := reconciler.store.PersistVerifiedSnapshot(ctx, claim, snapshot); err != nil {
		return ReconcileResult{}, fmt.Errorf("accept license snapshot: %w", err)
	}
	return ReconcileResult{Decision: decision, Snapshot: snapshot}, nil
}

func licenseIssueRequest(
	claim RefreshClaim,
	versions []ComponentVersion,
) IssueRequest {
	request := IssueRequest{
		OperationID: claim.OperationID, Binding: claim.Binding, Nonce: claim.Nonce,
		ComponentVersions: append([]ComponentVersion(nil), versions...),
	}
	if claim.Head != nil {
		request.CurrentSequence = claim.Head.Sequence
		request.CurrentDigest = claim.Head.Digest
	}
	return request
}

func (reconciler *Reconciler) persistFailure(
	ctx context.Context,
	claim RefreshClaim,
	kind FailureKind,
	cause error,
) error {
	recordErr := reconciler.store.StoreRefreshFailure(ctx, claim, kind)
	classification := ErrRefreshRetryable
	if kind == FailureConflict || kind == FailureRollback {
		classification = ErrRefreshFenced
	}
	return fmt.Errorf("reconcile license snapshot: %w",
		errors.Join(cause, recordErr, classification))
}
