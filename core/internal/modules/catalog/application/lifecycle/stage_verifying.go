package lifecycle

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) stageVerifying(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
) (ResumeResult, error) {
	if demoOperation(operation.Kind) {
		return application.stageDemoVerifying(ctx, scope, store, operation)
	}
	var authority PurgeAuthority
	var err error
	if operation.Kind == catalogdomain.OperationPurge {
		authority, _, err = application.revalidatePurgeAuthority(
			ctx, scope, store, operation,
		)
		if err != nil {
			return ResumeResult{}, err
		}
	}
	installation, err := store.GetInstallation(ctx, operation.Plan.PackageID)
	if err != nil {
		return ResumeResult{}, ErrNotFound
	}
	for _, pin := range operation.Plan.Components {
		owner, err := application.owners.Resolve(pin.Owner)
		if err != nil || owner == nil {
			return ResumeResult{}, ErrOwnerNotV2
		}
		component := componentowner.Component{ID: pin.ComponentID.String(), Owner: pin.Owner,
			Generation: pin.OwnerGeneration, Contract: pin.Contract, ContentSHA256: pin.ContentSHA256}
		readiness, err := owner.Readiness(ctx, scope, component)
		if err != nil || !readiness.WriterV2 || readiness.Component != component {
			return ResumeResult{}, ErrOwnerReceiptStale
		}
		if operation.Kind == catalogdomain.OperationPurge {
			command := componentowner.Command{OperationID: operation.ID.String(), Component: component,
				Action: "purge", ReferenceEpoch: installation.ReferenceEpoch,
				IdempotencyKey: operation.ID.String() + ":blockers"}
			receipt, err := owner.PurgeBlockers(ctx, scope, command)
			if err != nil || len(receipt.Blockers) != 0 || receipt.OperationID != operation.ID.String() ||
				receipt.ComponentID != pin.ComponentID.String() ||
				receipt.ReferenceEpoch != installation.ReferenceEpoch || !validStableID(receipt.ID) ||
				!validDigest(receipt.ZeroDigest) || receipt.CatalogVersion == "" ||
				receipt.CompletedAt.IsZero() || receipt.CompletedAt.Location() != time.UTC ||
				receipt.CompletedAt.Nanosecond()%1000 != 0 {
				return ResumeResult{}, ErrDependencyBlocked
			}
			if err = FinalizePurgeBlockerReceipt(ctx, store, PurgeBlockerReceipt{
				ReceiptID: receipt.ID, OperationID: receipt.OperationID, ComponentID: receipt.ComponentID,
				ReferenceEpoch: receipt.ReferenceEpoch, CatalogVersion: receipt.CatalogVersion,
				ZeroDigest: receipt.ZeroDigest, CompletedAt: receipt.CompletedAt,
			}); err != nil {
				return ResumeResult{}, err
			}
		}
	}
	if operation.Kind == catalogdomain.OperationPurge {
		retryAt, ready, finalizeErr := application.finalizePurgeOwners(
			ctx, scope, store, operation, authority,
		)
		if finalizeErr != nil {
			return ResumeResult{}, finalizeErr
		}
		if !ready {
			return ResumeResult{Operation: operation, NextAt: retryAt}, nil
		}
	}
	return application.advanceOperation(ctx, store, operation, catalogdomain.OperationActivating)
}

func validDigest(value string) bool { return sha256Pattern.MatchString(value) }
