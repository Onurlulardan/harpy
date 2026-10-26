package systemresources

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) presentMutation(
	ctx context.Context,
	scope platformtx.TenantScope,
	mutation publicresource.MutationResult,
) (MutationResult, error) {
	read, err := mutation.Read()
	if err != nil {
		return MutationResult{}, ErrContract
	}
	catalog, err := read.Catalog()
	if err != nil {
		return MutationResult{}, ErrContract
	}
	presentation, skip, err := application.recordPresentation(ctx, scope, catalog)
	if err != nil {
		return MutationResult{}, err
	}
	if skip {
		return newIAMOnlyMutationResult(mutation)
	}
	return NewMutationResult(mutation, presentation)
}
