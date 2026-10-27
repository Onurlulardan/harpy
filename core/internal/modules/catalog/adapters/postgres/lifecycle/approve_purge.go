package lifecycle

import (
	"context"
	_ "embed"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/create_purge_approval.sql
var createPurgeApprovalSQL string

//go:embed sql/tenant/get_purge_approval.sql
var getPurgeApprovalSQL string

func (storage *store) ApprovePurge(
	ctx context.Context,
	mutation catalogapp.PurgeApprovalMutation,
) (catalogapp.PurgeApprovalReceipt, bool, error) {
	value := mutation.Receipt
	if !value.Valid() || value.TenantID != storage.tenantID.String() ||
		value.ApprovedBy != storage.actorID.String() || !storage.usable(ctx, true) {
		return catalogapp.PurgeApprovalReceipt{}, false, catalogapp.ErrInvalid
	}
	schema, err := storage.currentSchema(ctx)
	if err != nil {
		return catalogapp.PurgeApprovalReceipt{}, false, unavailable()
	}
	stored, err := scanPurgeApproval(storage.queryRow(
		ctx, createPurgeApprovalSQL,
		schema, value.ID, value.TenantID, value.InstallationID, value.PackageID,
		value.OperationID, value.RequestedBy, value.ApprovedBy,
		value.BackupReceiptID, value.BackupSHA256, value.ConfirmationSlug,
		value.Reason, value.PurgeIdempotencyKey, value.RequestSHA256,
		value.ApprovalIdempotencyKey, value.StepUpAuthenticatedAt,
		value.ApprovedAt, value.ExpiresAt, value.ApprovalSHA256,
	))
	if err != nil || !stored.Valid() || stored.RequestSHA256 != value.RequestSHA256 ||
		stored.ApprovedBy != value.ApprovedBy ||
		stored.ApprovalIdempotencyKey != value.ApprovalIdempotencyKey {
		return catalogapp.PurgeApprovalReceipt{}, false, unavailable()
	}
	return stored, stored.ID != value.ID, nil
}

func (storage *store) ApprovalReceipt(
	ctx context.Context,
	receiptID uuid.UUID,
) (catalogapp.PurgeApprovalReceipt, error) {
	if receiptID.Version() != uuid.Version(7) || !storage.usable(ctx, false) {
		return catalogapp.PurgeApprovalReceipt{}, catalogapp.ErrInvalid
	}
	value, err := scanPurgeApproval(storage.queryRow(ctx, getPurgeApprovalSQL, receiptID))
	if err != nil {
		if storage.notFound(err) {
			return catalogapp.PurgeApprovalReceipt{}, catalogapp.ErrNotFound
		}
		return catalogapp.PurgeApprovalReceipt{}, unavailable()
	}
	if !value.Valid() || value.ID != receiptID.String() ||
		value.TenantID != storage.tenantID.String() {
		return catalogapp.PurgeApprovalReceipt{}, unavailable()
	}
	return value, nil
}

func scanPurgeApproval(source row) (catalogapp.PurgeApprovalReceipt, error) {
	var value catalogapp.PurgeApprovalReceipt
	var receiptID, tenantID, installationID, packageID, operationID uuid.UUID
	var requestedBy, approvedBy, backupID uuid.UUID
	var consumedOperationID *uuid.UUID
	var consumedAuthoritySHA256 *string
	var consumedAt *time.Time
	err := source.Scan(
		&receiptID, &tenantID, &installationID, &packageID, &operationID,
		&requestedBy, &approvedBy, &backupID, &value.BackupSHA256,
		&value.ConfirmationSlug, &value.Reason, &value.PurgeIdempotencyKey,
		&value.RequestSHA256, &value.ApprovalIdempotencyKey,
		&value.StepUpAuthenticatedAt, &value.ApprovedAt, &value.ExpiresAt,
		&value.State, &value.ApprovalSHA256, &consumedOperationID,
		&consumedAuthoritySHA256, &consumedAt,
	)
	if err != nil {
		return catalogapp.PurgeApprovalReceipt{}, err
	}
	value.ID, value.TenantID = receiptID.String(), tenantID.String()
	value.InstallationID, value.PackageID = installationID.String(), packageID.String()
	value.OperationID = operationID.String()
	value.RequestedBy, value.ApprovedBy = requestedBy.String(), approvedBy.String()
	value.BackupReceiptID = backupID.String()
	value.StepUpAuthenticatedAt = value.StepUpAuthenticatedAt.UTC()
	value.ApprovedAt, value.ExpiresAt = value.ApprovedAt.UTC(), value.ExpiresAt.UTC()
	if consumedOperationID != nil {
		if consumedAuthoritySHA256 == nil || consumedAt == nil {
			return catalogapp.PurgeApprovalReceipt{}, catalogapp.ErrUnavailable
		}
		value.ConsumedOperationID = consumedOperationID.String()
		value.ConsumedAuthoritySHA256 = *consumedAuthoritySHA256
		value.ConsumedAt = consumedAt.UTC()
	} else if consumedAuthoritySHA256 != nil || consumedAt != nil {
		return catalogapp.PurgeApprovalReceipt{}, catalogapp.ErrUnavailable
	}
	return value, nil
}

var _ catalogapp.PurgeApprovalStore = (*store)(nil)
