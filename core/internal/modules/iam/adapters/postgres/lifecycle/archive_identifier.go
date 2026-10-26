package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) ArchiveIdentifier(ctx context.Context,
	scope tenantScope,
	command lifecycle.RevisionCommand) (lifecycle.MutationResult, error) {
	return store.identifierRevision(ctx, scope, command, identifierRetireSQL)
}
