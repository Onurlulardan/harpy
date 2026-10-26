package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) SuspendAuthenticator(ctx context.Context,
	scope tenantScope,
	command lifecycle.RevisionCommand) (lifecycle.MutationResult, error) {
	return store.authenticatorRevision(ctx, scope, command, authenticatorSuspendSQL)
}
