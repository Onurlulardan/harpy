package iam

import (
	"context"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) GetSystemExport(
	ctx context.Context,
	input SystemExportReadInput,
) (view SystemExportView, err error) {
	if !systemTransferReadInputValid(ctx, input.Resource, input.ExportID) {
		return view, ErrSystemTransferInvalid
	}
	resource, _ := systemTransferResource(input.Resource)
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemExportGetOperation,
	)
	if err != nil {
		return view, err
	}
	err = runtime.systemTransferSnapshot(ctx, input.Access, values,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			store, bindErr := runtime.controls.BindControl(
				scoped, scope, systemtransfer.ControlReadAccess,
			)
			if bindErr != nil {
				return bindErr
			}
			record, readErr := systemtransfer.GetExport(scoped, store,
				systemTransferRead(input.Access, resource, input.ExportID))
			if readErr != nil {
				return readErr
			}
			view = systemExportView(record)
			return nil
		})
	return view, err
}
