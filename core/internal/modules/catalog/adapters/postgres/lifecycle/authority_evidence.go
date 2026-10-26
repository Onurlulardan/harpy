package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

//go:embed sql/tenant/record_purge_authority.sql
var recordPurgeAuthoritySQL string

func (storage *store) authorityEvidence(
	ctx context.Context,
	mutation catalogapp.StartMutation,
) error {
	if mutation.PurgeAuthority == nil {
		return nil
	}
	value := *mutation.PurgeAuthority
	if err := storage.approvalConsumption(ctx, value); err != nil {
		return err
	}
	var stored string
	err := storage.queryRow(ctx, recordPurgeAuthoritySQL,
		value.OperationID, value.TenantID, value.PackageID, value.PackageSlug,
		value.ActorID, value.RequestedBy, value.ApprovedBy, value.ApprovalReceiptID,
		value.ApprovalSHA256, value.BackupReceiptID, value.BackupState,
		value.BackupSHA256, value.ConfirmationSlug, value.Reason, value.IdempotencyKey,
		value.ReferenceEpoch, value.AuthorizationTokenVersion,
		value.AuthorizationVersion, value.AuthorizationClock,
		value.AuthorizationPermissionID, value.AuthorizationEvidenceSHA256,
		value.StepUpPurposeSHA256, value.StepUpAuthenticatedAt, value.ApprovalStepUpAt,
		value.ApprovedAt, value.ApprovalExpiresAt, value.RequestedAt,
		value.CoolingUntil, value.AuthoritySHA256,
	).Scan(&stored)
	if err != nil || stored != value.AuthoritySHA256 {
		return unavailable()
	}
	return nil
}
