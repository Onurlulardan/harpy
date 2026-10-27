package systemtransferartifact

import (
	"context"
	"io"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func (runtime *Runtime) OpenExactImportSource(
	ctx context.Context,
	source systemtransfer.ExactImportSource,
) (io.ReadCloser, error) {
	if runtime == nil || source.Validate() != nil {
		return nil, errInvalid
	}
	artifact := source.Artifact
	lease, err := runtime.decryptLocal(ctx, localOpenRequest{
		tenantID: source.TenantID, artifactID: source.SourceID,
		leaseOwner: source.ActorID, generation: 1, format: source.Format,
		object: artifactobject.Expected{
			Identity: artifactobject.Identity{Key: artifact.ObjectKey,
				Version: artifact.ObjectVersion},
			CiphertextBytes:  artifact.CiphertextBytes,
			CiphertextSHA256: artifact.CiphertextSHA256,
			HeaderSHA256:     artifact.HeaderSHA256,
		},
		keyID: artifact.KeyID, wrappedDEK: artifact.WrappedDEK,
		wrapNonce: artifact.WrapNonce, expectedPlaintextBytes: artifact.PlaintextBytes,
		expectedPlaintextSHA256: &artifact.PlaintextSHA256,
		keys:                    runtime.importKeys, objects: runtime.imports,
	})
	if err != nil {
		return nil, err
	}
	return openDiscardingReader(lease)
}

var _ systemtransfer.ExactImportSourceOpener = (*Runtime)(nil)
