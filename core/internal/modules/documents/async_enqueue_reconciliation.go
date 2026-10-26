package documents

import (
	"context"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *APIAsyncPort) EnqueueReconciliation(ctx context.Context,
	scope platformtx.TenantScope, jobID, uploadID foundation.ID, generation uint64,
	runAt time.Time, payload []byte,
) error {
	reference, digest, err := jobcontract.UploadReconcile(
		jobcontract.ID(scope.TenantID()), jobcontract.ID(uploadID), generation)
	if err != nil {
		return ErrDocumentsRuntime
	}
	key := uploadReconcileKey(uploadID.String(), generation)
	return port.enqueue(ctx, scope, jobID,
		asyncmanifestv1.DocumentsUploadReconcileJobKind, reference, digest, key,
		runAt, payload)
}
