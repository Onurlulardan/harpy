package iam

import (
	"context"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) CreateSystemImportSource(
	ctx context.Context,
	input SystemImportSourceInput,
) (view SystemImportSourceView, err error) {
	if ctx == nil || ctx.Err() != nil || !input.Resource.Valid() || input.Body == nil ||
		(input.Format != "csv" && input.Format != "jsonl") {
		return view, ErrSystemTransferInvalid
	}
	resource, err := systemTransferResource(input.Resource)
	if err != nil {
		return view, err
	}
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemImportSourceOperation,
	)
	if err != nil {
		return view, err
	}
	sourceID, err := runtime.newV7()
	if err != nil {
		return view, err
	}
	receiptID, err := runtime.newV7()
	if err != nil {
		return view, err
	}
	keyDigest, err := systemTransferKeySHA256(input.IdempotencyKey)
	if err != nil {
		return view, err
	}
	now, err := runtime.now()
	if err != nil {
		return view, err
	}
	record, err := runtime.artifacts.SealImportSource(ctx,
		systemtransfer.SourceSealRequest{
			TenantID: input.Access.TenantID(), ActorID: input.Access.PrincipalID(),
			SourceID: sourceID, Resource: resource, Format: input.Format,
			IdempotencySHA256: keyDigest, UploadedAt: now, Body: input.Body,
			MaximumBytes: SystemTransferMaximumUploadBytes,
		},
		func(callbackCtx context.Context,
			artifact systemtransfer.SourceArtifact,
		) (stored systemtransfer.SourceRecord, callbackErr error) {
			callbackErr = runtime.systemTransferCommand(callbackCtx, input.Access, values,
				SystemImportSourceOperation,
				func(scoped context.Context, scope platformtx.TenantScope) error {
					store, bindErr := runtime.controls.BindControl(
						scoped, scope, systemtransfer.ControlCommandAccess,
					)
					if bindErr != nil {
						return bindErr
					}
					stored, bindErr = store.RegisterSource(scoped, systemtransfer.SourceWrite{
						Evidence: systemTransferEvidence(input.Access, resource),
						SourceID: sourceID, ReceiptID: receiptID,
						IdempotencyKeySHA256: keyDigest, Format: input.Format,
						Artifact: artifact,
					})
					return bindErr
				})
			return stored, callbackErr
		})
	if err != nil {
		return view, systemTransferError(err)
	}
	return systemImportSourceView(record), nil
}
