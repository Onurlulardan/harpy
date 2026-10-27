package iam

import (
	"context"
	"encoding/hex"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) SubmitSystemImport(
	ctx context.Context,
	input SystemImportSubmitInput,
) (view SystemImportView, err error) {
	if !systemTransferReadInputValid(ctx, input.Resource, input.SourceID) ||
		(input.DuplicateAction != "skip" && input.DuplicateAction != "update") ||
		len(input.MappingDocument) < 2 {
		return view, ErrSystemTransferInvalid
	}
	resource, _ := systemTransferResource(input.Resource)
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemImportSubmitOperation,
	)
	if err != nil {
		return view, err
	}
	var prepared systemTransferPreparedMapping
	err = runtime.systemTransferSnapshot(ctx, input.Access, values,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var prepareErr error
			prepared, prepareErr = runtime.prepareSystemImportMapping(
				scoped, scope, input.Access, resource, input.MappingDocument,
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
	importID, jobID, receiptID, err := runtime.threeV7()
	if err != nil {
		return view, err
	}
	keyDigest, err := systemTransferKeySHA256(input.IdempotencyKey)
	if err != nil {
		return view, err
	}
	requestDigest, err := systemTransferRequestSHA256(struct {
		Resource, SourceID, DuplicateAction, MappingSHA256 string
	}{string(input.Resource), input.SourceID.String(), input.DuplicateAction,
		hex.EncodeToString(prepared.digest[:])})
	if err != nil {
		return view, err
	}
	err = runtime.systemTransferCommand(ctx, input.Access, values,
		SystemImportSubmitOperation,
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
			record, submitErr := systemtransfer.ImportRequest(scoped, store,
				systemtransfer.ImportWrite{
					Evidence: systemTransferEvidence(input.Access, resource),
					ImportID: importID, JobID: jobID, ReceiptID: receiptID,
					SourceID: input.SourceID, CatalogVersionID: prepared.catalog.versionID,
					CatalogSHA256:          prepared.catalog.sha256,
					IdentityManifestSHA256: prepared.manifestDigest,
					RequestSHA256:          requestDigest, IdempotencyKeySHA256: keyDigest,
					MappingSHA256: prepared.digest, DuplicateAction: input.DuplicateAction,
					MappingDocument: prepared.document, FieldCount: prepared.fieldCount,
					CreatedAt: now,
				})
			if submitErr != nil {
				return submitErr
			}
			if !record.Replay {
				if enqueueErr := runtime.enqueueSystemTransfer(scoped, scope, input.Resource,
					record.ImportID, record.JobID, SystemImportJobKind, now); enqueueErr != nil {
					return enqueueErr
				}
			}
			view = systemImportView(record)
			return nil
		})
	return view, err
}
