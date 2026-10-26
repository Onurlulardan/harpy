package iam

import (
	"context"
	"encoding/hex"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) SubmitSystemExport(
	ctx context.Context,
	input SystemExportSubmitInput,
) (view SystemExportView, err error) {
	if ctx == nil || ctx.Err() != nil || !input.Resource.Valid() ||
		(input.Format != "csv" && input.Format != "jsonl" && input.Format != "xlsx") ||
		len(input.QueryDocument) < 2 {
		return view, ErrSystemTransferInvalid
	}
	resource, err := systemTransferResource(input.Resource)
	if err != nil {
		return view, err
	}
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemExportSubmitOperation,
	)
	if err != nil {
		return view, err
	}
	var prepared systemTransferPreparedQuery
	err = runtime.systemTransferSnapshot(ctx, input.Access, values,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var prepareErr error
			prepared, prepareErr = runtime.prepareSystemExportQuery(
				scoped, scope, input.Resource, input.QueryDocument,
			)
			return prepareErr
		})
	if err != nil {
		return view, err
	}
	now, err := runtime.now()
	if err != nil {
		return view, err
	}
	exportID, jobID, receiptID, err := runtime.threeV7()
	if err != nil {
		return view, err
	}
	keyDigest, err := systemTransferKeySHA256(input.IdempotencyKey)
	if err != nil {
		return view, err
	}
	requestDigest, err := systemTransferRequestSHA256(struct {
		Resource, Format, QuerySHA256 string
	}{string(input.Resource), input.Format, hex.EncodeToString(prepared.digest[:])})
	if err != nil {
		return view, err
	}
	err = runtime.systemTransferCommand(ctx, input.Access, values,
		SystemExportSubmitOperation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			current, currentErr := runtime.currentSystemTransferPin(scoped, scope, input.Resource)
			if currentErr != nil || !prepared.catalog.same(current) {
				return ErrSystemTransferConflict
			}
			store, bindErr := runtime.controls.BindControl(
				scoped, scope, systemtransfer.ControlCommandAccess,
			)
			if bindErr != nil {
				return bindErr
			}
			record, submitErr := systemtransfer.ExportRequest(scoped, store,
				systemtransfer.ExportWrite{
					Evidence: systemTransferEvidence(input.Access, resource),
					ExportID: exportID, JobID: jobID, ReceiptID: receiptID,
					CatalogVersionID: prepared.catalog.versionID,
					CatalogSHA256:    prepared.catalog.sha256,
					QuerySHA256:      prepared.digest, RequestSHA256: requestDigest,
					IdempotencyKeySHA256: keyDigest, QueryDocument: prepared.document,
					Format: input.Format, CreatedAt: now,
				})
			if submitErr != nil {
				return submitErr
			}
			if !record.Replay {
				if enqueueErr := runtime.enqueueSystemTransfer(scoped, scope, input.Resource,
					record.ExportID, record.JobID, SystemExportJobKind, now); enqueueErr != nil {
					return enqueueErr
				}
			}
			view = systemExportView(record)
			return nil
		})
	return view, err
}
