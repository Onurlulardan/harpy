package iam

import (
	"context"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) ReportSystemImport(
	ctx context.Context,
	input SystemImportReportInput,
) (view SystemImportReportView, err error) {
	if !systemTransferReadInputValid(ctx, input.Resource, input.ImportID) ||
		input.AfterRow < 0 || input.AfterRow > 1_000_000 ||
		input.Limit < 1 || input.Limit > 100 {
		return view, ErrSystemTransferInvalid
	}
	resource, _ := systemTransferResource(input.Resource)
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemImportReportOperation,
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
			page, readErr := systemtransfer.ListReceipts(scoped, store,
				systemTransferRead(input.Access, resource, input.ImportID),
				input.AfterRow, input.Limit)
			if readErr != nil {
				return readErr
			}
			view = systemImportReportView(input.ImportID, page)
			return nil
		})
	return view, err
}
