package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) presentRead(
	ctx context.Context,
	scope platformtx.TenantScope,
	read publicresource.ReadResult,
) (ReadResult, error) {
	catalog, err := read.Catalog()
	if err != nil {
		return ReadResult{}, ErrContract
	}
	presentation, skip, err := application.recordPresentation(ctx, scope, catalog)
	if err != nil {
		return ReadResult{}, err
	}
	if skip {
		return newIAMOnlyReadResult(read)
	}
	return NewReadResult(read, presentation)
}
