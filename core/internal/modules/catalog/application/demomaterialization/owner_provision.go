package demomaterialization

import (
	"context"
	"fmt"
	"time"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) Provision(
	ctx context.Context,
	scope platformtx.TenantScope,
	request contract.ProvisionRequest,
) (contract.ProvisionReceipt, error) {
	verified, err := verifyRequest(request)
	if err != nil {
		return contract.ProvisionReceipt{}, err
	}
	store, err := application.bind(ctx, scope, request.Component)
	if err != nil {
		return contract.ProvisionReceipt{}, err
	}
	now := application.clock.Now().UTC().Truncate(time.Microsecond)
	receiptID, idErr := application.ids.NewUUIDv7()
	requestSHA256, digestErr := projectionDigest(verified)
	if idErr != nil || digestErr != nil || receiptID.Version() != 7 ||
		now.IsZero() || now.Year() < 2020 {
		return contract.ProvisionReceipt{}, ErrUnavailable
	}
	projection := Projection{
		ProposedReceiptID: receiptID,
		PackageID:         verified.packageID, PackageSlug: request.PackageSlug,
		PackageVersion: request.PackageVersion, ReleaseID: verified.releaseID,
		Component: request.Component, Materialization: verified.materialization.Clone(),
		IdempotencyKey: request.IdempotencyKey,
		RequestSHA256:  requestSHA256, OccurredAt: now,
	}
	snapshot, err := store.Project(ctx, projection)
	if err != nil {
		return contract.ProvisionReceipt{}, fmt.Errorf(
			"project Catalog demo component: %w", err,
		)
	}
	if !snapshotMatches(snapshot, verified, requestSHA256) {
		return contract.ProvisionReceipt{}, ErrConflict
	}
	return receiptFromSnapshot(snapshot), nil
}

func receiptFromSnapshot(snapshot Snapshot) contract.ProvisionReceipt {
	return contract.ProvisionReceipt{
		ID: snapshot.ReceiptID.String(), PackageID: snapshot.PackageID.String(),
		PackageSlug: snapshot.PackageSlug, PackageVersion: snapshot.PackageVersion,
		ReleaseID: snapshot.ReleaseID.String(), Component: snapshot.Component,
		State: snapshot.State, Ready: snapshot.Ready, RetryAt: snapshot.RetryAt,
		CompletedAt: snapshot.CompletedAt,
	}
}

func (application *Application) bind(
	ctx context.Context,
	scope platformtx.TenantScope,
	component contract.Component,
) (Store, error) {
	if application == nil || ctx == nil || ctx.Err() != nil ||
		!platformtx.ContextOwnsScope(ctx, scope) || platformtx.NetworkAllowed(ctx) ||
		scope.Mode() != platformtx.Command || scope.OperationID() != ProvisionOperation ||
		platformtx.RequireOwner(scope, owner) != nil {
		return nil, ErrUnavailable
	}
	authority, err := platformtx.MintOwnerAuthority(scope, owner)
	if err != nil || authority.Generation() != component.Generation {
		return nil, ErrUnavailable
	}
	store, err := application.binder.Bind(ctx, scope)
	if err != nil || nilDependency(store) {
		return nil, ErrUnavailable
	}
	return store, nil
}
