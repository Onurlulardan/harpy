package documents

import (
	"context"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *APIAsyncPort) Enqueue(ctx context.Context, scope platformtx.TenantScope,
	jobID, archiveID foundation.ID, payload []byte,
) error {
	reference, digest, err := jobcontract.Archive(jobcontract.ID(scope.TenantID()),
		jobcontract.ID(archiveID))
	if err != nil {
		return ErrDocumentsRuntime
	}
	key := "documents:archive:" + archiveID.String()
	return port.enqueue(ctx, scope, jobID,
		asyncmanifestv1.DocumentsArchiveBuildJobKind, reference, digest, key,
		port.clock.Now(), payload)
}
