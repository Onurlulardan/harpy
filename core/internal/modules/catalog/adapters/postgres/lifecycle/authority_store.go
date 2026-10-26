package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/get_purge_authority.sql
var getPurgeAuthoritySQL string

func (storage *store) AuthorityForPurge(
	ctx context.Context,
	operationID uuid.UUID,
) (catalogapp.PurgeAuthority, error) {
	if operationID == uuid.Nil || !storage.usable(ctx, false) {
		return catalogapp.PurgeAuthority{}, catalogapp.ErrInvalid
	}
	var value catalogapp.PurgeAuthority
	var tenantID, packageID, actorID, requestedBy, approvedBy, approvalID, backupID uuid.UUID
	var storedOperationID uuid.UUID
	err := storage.queryRow(ctx, getPurgeAuthoritySQL, operationID).Scan(
		&tenantID, &packageID, &value.PackageSlug, &storedOperationID, &actorID,
		&requestedBy, &approvedBy, &backupID, &value.BackupState,
		&approvalID, &value.ApprovalSHA256, &value.BackupSHA256,
		&value.ConfirmationSlug, &value.Reason,
		&value.IdempotencyKey, &value.ReferenceEpoch,
		&value.StepUpAuthenticatedAt,
		&value.AuthorizationTokenVersion, &value.AuthorizationVersion,
		&value.AuthorizationClock, &value.AuthorizationPermissionID,
		&value.AuthorizationEvidenceSHA256, &value.StepUpPurposeSHA256,
		&value.ApprovalStepUpAt,
		&value.ApprovedAt, &value.ApprovalExpiresAt, &value.RequestedAt, &value.CoolingUntil,
		&value.AuthoritySHA256,
	)
	if err != nil {
		if storage.notFound(err) {
			return catalogapp.PurgeAuthority{}, catalogapp.ErrNotFound
		}
		return catalogapp.PurgeAuthority{}, unavailable()
	}
	value.TenantID, value.PackageID = tenantID.String(), packageID.String()
	value.OperationID, value.ActorID = storedOperationID.String(), actorID.String()
	value.RequestedBy, value.ApprovedBy = requestedBy.String(), approvedBy.String()
	value.ApprovalReceiptID = approvalID.String()
	value.BackupReceiptID = backupID.String()
	value.StepUpAuthenticatedAt = value.StepUpAuthenticatedAt.UTC()
	value.ApprovalStepUpAt = value.ApprovalStepUpAt.UTC()
	value.ApprovedAt, value.ApprovalExpiresAt = value.ApprovedAt.UTC(), value.ApprovalExpiresAt.UTC()
	value.RequestedAt, value.CoolingUntil = value.RequestedAt.UTC(), value.CoolingUntil.UTC()
	if storedOperationID != operationID || !value.Valid() {
		return catalogapp.PurgeAuthority{}, unavailable()
	}
	return value, nil
}

var _ catalogapp.PurgeAuthorityStore = (*store)(nil)
