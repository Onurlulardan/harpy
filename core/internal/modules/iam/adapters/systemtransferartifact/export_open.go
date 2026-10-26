package systemtransferartifact

import (
	"context"
	"io"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (runtime *Runtime) OpenExport(
	ctx context.Context,
	request systemtransfer.ExportOpenRequest,
) (io.ReadCloser, error) {
	if runtime == nil || !validExportOpen(ctx, request) {
		return nil, errInvalid
	}
	artifact := request.Artifact
	lease, err := runtime.decryptLocal(ctx, localOpenRequest{
		tenantID: request.TenantID, artifactID: request.ExportID,
		leaseOwner: request.TenantID, generation: artifact.AttemptGeneration,
		format: artifact.Format,
		object: artifactobject.Expected{
			Identity: artifactobject.Identity{Key: artifact.ObjectKey,
				Version: artifact.ObjectVersion},
			CiphertextBytes:  artifact.CiphertextBytes,
			CiphertextSHA256: artifact.CiphertextSHA256,
			HeaderSHA256:     artifact.HeaderSHA256,
		},
		keyID: artifact.KeyID, wrappedDEK: artifact.WrappedDEK,
		wrapNonce: artifact.WrapNonce, expectedPlaintextBytes: artifact.Bytes,
		keys: runtime.exportKeys, objects: runtime.exports,
	})
	if err != nil {
		return nil, err
	}
	return openDiscardingReader(lease)
}

func validExportOpen(ctx context.Context,
	request systemtransfer.ExportOpenRequest) bool {
	artifact := request.Artifact
	return runtimeInputContext(ctx) && request.TenantID != [16]byte{} &&
		request.ExportID.Version() == 7 && request.Resource.Valid() &&
		artifact.ExportID == request.ExportID && artifact.Resource == string(request.Resource) &&
		artifact.State == "ready" && artifact.ReadyAt != nil && artifact.ExpiresAt != nil &&
		(artifact.Format == "csv" || artifact.Format == "jsonl" || artifact.Format == "xlsx") &&
		artifact.Bytes > 0 && artifact.ObjectKey != "" && artifact.ObjectVersion != "" &&
		artifact.KeyID != "" && artifact.HeaderSHA256 != [32]byte{} &&
		artifact.CiphertextSHA256 != [32]byte{} && artifact.CiphertextBytes > 0 &&
		len(artifact.WrappedDEK) >= 32 && artifact.WrapNonce != [12]byte{} &&
		artifact.AttemptGeneration >= 1 && artifact.AttemptGeneration <= 3
}

var _ systemtransfer.ArtifactRuntime = (*Runtime)(nil)
