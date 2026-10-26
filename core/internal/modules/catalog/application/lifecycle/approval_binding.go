package lifecycle

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (application *Application) approvalForPurge(
	ctx context.Context,
	store Store,
	identity ScopeIdentity,
	installationID string,
	request PurgeRequest,
	backup BackupEvidence,
	now time.Time,
) (PurgeApprovalReceipt, error) {
	reader, ok := store.(PurgeApprovalStore)
	if !ok || reader == nil {
		return PurgeApprovalReceipt{}, ErrUnavailable
	}
	receipt, err := reader.ApprovalReceipt(ctx, request.ApprovalReceiptID)
	if err != nil {
		return PurgeApprovalReceipt{}, err
	}
	installationUUID, err := uuid.Parse(installationID)
	if err != nil {
		return PurgeApprovalReceipt{}, ErrUnavailable
	}
	requestSHA, err := approvalRequestSHA(
		identity.TenantID, request.PackageID,
		PurgeApprovalRequest{
			InstallationID: installationUUID, OperationID: request.OperationID,
			BackupReceiptID: request.BackupReceiptID, RequestedBy: request.RequestedBy,
			Reason: request.Reason, ConfirmationSlug: request.ConfirmationSlug,
			PurgeIdempotencyKey: request.IdempotencyKey,
		},
	)
	if err != nil {
		return PurgeApprovalReceipt{}, err
	}
	if !receipt.Valid() || receipt.State != purgeApprovalAvailable ||
		receipt.ID != request.ApprovalReceiptID.String() ||
		receipt.TenantID != identity.TenantID.String() ||
		receipt.InstallationID != installationID ||
		receipt.PackageID != request.PackageID.String() ||
		receipt.OperationID != request.OperationID.String() ||
		receipt.RequestedBy != request.RequestedBy.String() ||
		receipt.ApprovedBy == identity.ActorID.String() ||
		receipt.BackupReceiptID != request.BackupReceiptID.String() ||
		receipt.BackupSHA256 != backup.SHA256 ||
		receipt.ConfirmationSlug != request.ConfirmationSlug ||
		receipt.Reason != request.Reason ||
		receipt.PurgeIdempotencyKey != request.IdempotencyKey ||
		receipt.RequestSHA256 != requestSHA || now.After(receipt.ExpiresAt) {
		return PurgeApprovalReceipt{}, ErrForbidden
	}
	return receipt, nil
}
