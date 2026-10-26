package serviceprincipal

import (
	"context"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func (store *commandStore) Enable(
	ctx context.Context, scope tenantScope, command serviceapp.StateCommand,
) (serviceapp.MutationResult, error) {
	return store.writeState(ctx, scope, command, enableSQL)
}
