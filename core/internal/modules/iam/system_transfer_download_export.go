package iam

import (
	"context"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *SystemTransferRuntime) DownloadSystemExport(
	ctx context.Context,
	input SystemExportReadInput,
) (content SystemExportContent, err error) {
	if !systemTransferReadInputValid(ctx, input.Resource, input.ExportID) {
		return content, ErrSystemTransferInvalid
	}
	resource, _ := systemTransferResource(input.Resource)
	values, err := runtime.resolveSystemTransfer(
		ctx, input.Access, input.Resource, SystemExportDownloadOperation,
	)
	if err != nil {
		return content, err
	}
	now, err := runtime.now()
	if err != nil {
		return content, err
	}
	var artifact systemtransfer.ExportArtifact
	err = runtime.systemTransferSnapshot(ctx, input.Access, values,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			store, bindErr := runtime.controls.BindControl(
				scoped, scope, systemtransfer.ControlReadAccess,
			)
			if bindErr != nil {
				return bindErr
			}
			var readErr error
			artifact, readErr = systemtransfer.DownloadExportArtifact(scoped, store,
				systemTransferRead(input.Access, resource, input.ExportID), now)
			return readErr
		})
	if err != nil {
		return content, err
	}
	body, err := runtime.artifacts.OpenExport(ctx, systemtransfer.ExportOpenRequest{
		TenantID: input.Access.TenantID(), Resource: resource,
		ExportID: input.ExportID, Artifact: artifact,
	})
	if err != nil || body == nil {
		return content, ErrSystemTransferUnavailable
	}
	name, contentType, err := systemExportContentMetadata(input.Resource, artifact)
	if err != nil {
		_ = body.Close()
		return content, err
	}
	return SystemExportContent{
		Body: body, Size: artifact.Bytes, Name: name, ContentType: contentType,
	}, nil
}
