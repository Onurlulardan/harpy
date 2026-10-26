package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

//go:embed sql/tenant/consume_purge_approval.sql
var consumePurgeApprovalSQL string

func (storage *store) approvalConsumption(
	ctx context.Context,
	value catalogapp.PurgeAuthority,
) error {
	schema, err := storage.currentSchema(ctx)
	if err != nil {
		return unavailable()
	}
	var stored string
	err = storage.queryRow(
		ctx, consumePurgeApprovalSQL,
		schema, value.ApprovalReceiptID, value.TenantID, value.PackageID, value.OperationID,
		value.RequestedBy, value.ApprovedBy, value.BackupReceiptID,
		value.ConfirmationSlug, value.Reason, value.IdempotencyKey,
		value.ApprovalSHA256, value.AuthoritySHA256, value.RequestedAt,
	).Scan(&stored)
	if err != nil || stored != value.ApprovalSHA256 {
		return unavailable()
	}
	return nil
}
