package documents

import (
	"context"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/api/jobcontract"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (port *APIAsyncPort) EnqueuePurge(ctx context.Context,
	scope platformtx.TenantScope, jobID, itemID foundation.ID,
	runAt time.Time, payload []byte,
) error {
	reference, digest, err := jobcontract.Purge(
		jobcontract.ID(scope.TenantID()), jobcontract.ID(itemID))
	if err != nil {
		return ErrDocumentsRuntime
	}
	key := "documents:forms-cleanup:purge:" + itemID.String()
	return port.enqueue(ctx, scope, jobID,
		asyncmanifestv1.DocumentsItemPurgeJobKind, reference, digest, key,
		runAt, payload)
}
