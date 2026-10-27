package lifecycle

import (
	"context"
	"time"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) revalidatePurgeAuthority(
	ctx context.Context,
	scope platformtx.TenantScope,
	store Store,
	operation catalogdomain.Operation,
) (PurgeAuthority, time.Time, error) {
	authorityStore, ok := store.(PurgeAuthorityStore)
	if !ok || authorityStore == nil {
		return PurgeAuthority{}, time.Time{}, ErrUnavailable
	}
	operationID, err := uuidFromStable(operation.ID)
	if err != nil {
		return PurgeAuthority{}, time.Time{}, err
	}
	authority, err := authorityStore.AuthorityForPurge(ctx, operationID)
	if err != nil || !authority.Valid() ||
		authority.TenantID != scope.TenantID().String() ||
		authority.ActorID != scope.ActorID().String() ||
		authority.OperationID != operation.ID.String() ||
		authority.PackageID != operation.Plan.PackageID.String() ||
		authority.PackageSlug != operation.Plan.Slug {
		return PurgeAuthority{}, time.Time{}, ErrUnavailable
	}
	installation, err := store.GetInstallation(ctx, operation.Plan.PackageID)
	if err != nil || installation.State != catalogdomain.InstallationPurgePending ||
		installation.ReferenceEpoch != authority.ReferenceEpoch {
		return PurgeAuthority{}, time.Time{}, ErrConflict
	}
	now, err := application.now()
	if err != nil || now.Before(authority.RequestedAt) {
		return PurgeAuthority{}, time.Time{}, ErrUnavailable
	}
	if now.Before(authority.CoolingUntil) {
		return authority, now, nil
	}
	requirement := BackupRequirement{
		TenantID: authority.TenantID, PackageID: authority.PackageID,
		CatalogOperationID: authority.OperationID,
		OwnerGenerations:   ownerGenerations(operation.Plan),
		Now:                now, MinimumExpiry: now.Add(90 * 24 * time.Hour),
	}
	evidence, err := application.backup.Verify(
		ctx, scope, authority.BackupReceiptID, requirement,
	)
	if err != nil || evidence.ReceiptID != authority.BackupReceiptID ||
		evidence.TenantID != authority.TenantID || evidence.PackageID != authority.PackageID ||
		evidence.CatalogOperationID != authority.OperationID ||
		evidence.State != catalogcomponentv2.BackupReady ||
		evidence.SHA256 != authority.BackupSHA256 {
		return PurgeAuthority{}, time.Time{}, ErrBackupBlocked
	}
	return authority, now, nil
}

func ownerGenerations(plan catalogdomain.Plan) []BackupGeneration {
	values := make([]BackupGeneration, 0, len(plan.Components))
	seen := make(map[string]bool, len(plan.Components))
	for _, component := range plan.Components {
		if !seen[component.Owner] {
			seen[component.Owner] = true
			values = append(values, BackupGeneration{
				Owner: component.Owner, Generation: component.OwnerGeneration,
			})
		}
	}
	return values
}
