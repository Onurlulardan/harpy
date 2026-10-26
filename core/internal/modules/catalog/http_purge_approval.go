package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) ApprovePurge(
	ctx context.Context,
	access cataloghttp.Access,
	input cataloghttp.PurgeApprovalInput,
) (cataloghttp.PurgeApproval, error) {
	if access.StepUpAt().IsZero() || !input.StepUpAt.Equal(access.StepUpAt()) {
		return cataloghttp.PurgeApproval{}, cataloghttp.ErrForbidden
	}
	var result cataloghttp.PurgeApproval
	err := runtime.lifecycleCommand(
		ctx, access, cataloghttp.PermissionPurge,
		"catalog.installation.purge.approve",
		func(scoped context.Context, scope platformtx.TenantScope) error {
			receipt, replayed, err := runtime.lifecycle.ApprovePurge(
				scoped, scope, lifecycleapp.PurgeApprovalRequest{
					InstallationID: input.InstallationID, OperationID: input.OperationID,
					BackupReceiptID: input.BackupReceiptID, RequestedBy: input.RequestedBy,
					StepUpAtUnix: input.StepUpAt.Unix(), Reason: input.Reason,
					ConfirmationSlug:       input.ConfirmationSlug,
					PurgeIdempotencyKey:    input.PurgeIdempotencyKey,
					ApprovalIdempotencyKey: input.ApprovalIdempotencyKey,
				},
			)
			if err == nil {
				result = cataloghttp.RepresentPurgeApproval(receipt, replayed)
			}
			return err
		},
	)
	return result, err
}
