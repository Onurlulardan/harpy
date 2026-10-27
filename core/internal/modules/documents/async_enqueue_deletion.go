package documents

import (
	"context"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *APIAsyncPort) EnqueueDeletion(ctx context.Context,
	scope platformtx.TenantScope, jobID, deletionID foundation.ID,
	payload []byte,
) error {
	reference, digest, err := jobcontract.Deletion(jobcontract.ID(scope.TenantID()),
		jobcontract.ID(deletionID))
	if err != nil {
		return ErrDocumentsRuntime
	}
	key := "documents:blob-delete:" + deletionID.String()
	return port.enqueue(ctx, scope, jobID,
		asyncmanifestv1.DocumentsBlobDeleteJobKind, reference, digest, key,
		port.clock.Now(), payload)
}
