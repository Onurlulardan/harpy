package recovery

import (
	"context"

	"github.com/google/uuid"
)

func (service *Service) RecordCopy(
	ctx context.Context,
	recoverySetID uuid.UUID,
	item InventoryItem,
	targetVersion string,
) error {
	if !service.usable(ctx) || recoverySetID == uuid.Nil ||
		!validInventoryItem(item) || targetVersion == "" {
		return ErrInvalid
	}
	key := recoveryObjectKey(
		service.identity.EnvironmentID, recoverySetID,
		service.identity.TenantID, item.VersionID,
	)
	return service.store.RecordCopy(ctx, CopyReceipt{
		RecoverySetID: recoverySetID, Item: item,
		TargetObjectKey: key, TargetObjectVersion: targetVersion,
	})
}

func recoveryObjectKey(
	environmentID uuid.UUID,
	recoverySetID uuid.UUID,
	tenantID uuid.UUID,
	versionID uuid.UUID,
) string {
	return "environments/" + environmentID.String() +
		"/recovery-sets/" + recoverySetID.String() +
		"/documents/tenants/" + tenantID.String() +
		"/versions/" + versionID.String() + ".fcdoc1"
}
