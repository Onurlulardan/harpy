package iam

import (
	"context"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) GetSystemImport(
	ctx context.Context,
	input SystemImportReadInput,
) (view SystemImportView, err error) {
	if !systemTransferReadInputValid(ctx, input.Resource, input.ImportID) {
		return view, ErrSystemTransferInvalid
	}
	resource, _ := systemTransferResource(input.Resource)
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemImportGetOperation,
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
			record, readErr := systemtransfer.ResolveImport(scoped, store,
				systemTransferRead(input.Access, resource, input.ImportID))
			if readErr != nil {
				return readErr
			}
			view = systemImportView(record)
			return nil
		})
	return view, err
}
