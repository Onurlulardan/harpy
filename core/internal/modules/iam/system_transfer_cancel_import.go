package iam

import (
	"context"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) CancelSystemImport(
	ctx context.Context,
	input SystemImportReadInput,
) (view SystemImportView, err error) {
	if !systemTransferReadInputValid(ctx, input.Resource, input.ImportID) {
		return view, ErrSystemTransferInvalid
	}
	resource, _ := systemTransferResource(input.Resource)
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemImportCancelOperation,
	)
	if err != nil {
		return view, err
	}
	now, err := runtime.now()
	if err != nil {
		return view, err
	}
	err = runtime.systemTransferCommand(ctx, input.Access, values,
		SystemImportCancelOperation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			store, bindErr := runtime.controls.BindControl(
				scoped, scope, systemtransfer.ControlCommandAccess,
			)
			if bindErr != nil {
				return bindErr
			}
			record, cancelErr := systemtransfer.DeleteImport(scoped, store,
				systemTransferRead(input.Access, resource, input.ImportID), now)
			if cancelErr != nil {
				return cancelErr
			}
			if err := runtime.cancelSystemTransferJob(scoped, scope, record.JobID,
				now, "IAM_SYSTEM_IMPORT_CANCELLED"); err != nil {
				return err
			}
			view = systemImportView(record)
			return nil
		})
	return view, err
}
