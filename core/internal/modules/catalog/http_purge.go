package catalog

import (
	"context"

	cataloghttp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/http"
	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (runtime *HTTPRuntime) Purge(ctx context.Context, access cataloghttp.Access,
	input cataloghttp.PurgeInput,
) (cataloghttp.Operation, error) {
	if access.StepUpAt().IsZero() || !input.StepUpAt.Equal(access.StepUpAt()) {
		return cataloghttp.Operation{}, cataloghttp.ErrForbidden
	}
	var result cataloghttp.Operation
	err := runtime.lifecycleCommand(ctx, access, cataloghttp.PermissionPurge,
		"catalog.installation.purge", func(scoped context.Context, scope platformtx.TenantScope) error {
			installation, err := runtime.lifecycle.CommandInstallation(
				scoped, scope, input.InstallationID, lifecycleapp.ActionPurge)
			if err != nil {
				return err
			}
			packageID, err := uuid.Parse(installation.PackageID.String())
			if err != nil {
				return cataloghttp.ErrUnavailable
			}
			handle, err := runtime.lifecycle.Purge(scoped, scope, lifecycleapp.PurgeRequest{
				OperationID: input.OperationID, PackageID: packageID,
				BackupReceiptID:   input.BackupReceiptID,
				ApprovalReceiptID: input.ApprovalReceiptID,
				RequestedBy:       access.PrincipalID(), StepUpAtUnix: input.StepUpAt.Unix(),
				Reason: input.Reason, ConfirmationSlug: input.ConfirmationSlug,
				IdempotencyKey: input.IdempotencyKey})
			if err == nil {
				result = cataloghttp.RepresentHandle(handle)
			}
			return err
		})
	return result, err
}
