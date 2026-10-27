package serviceprincipal

import (
	"context"

	serviceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/serviceprincipal"
)

func (store *commandStore) Retire(
	ctx context.Context, scope tenantScope, command serviceapp.StateCommand,
) (serviceapp.MutationResult, error) {
	return store.writeState(ctx, scope, command, retireSQL)
}
