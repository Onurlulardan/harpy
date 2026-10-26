package systemtransfer

import (
	"context"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

type WorkerStore struct {
	bindOwned func(tenantScope) (*workerBound, error)
}

type workerBound struct {
	database boundDatabase
	notFound func(error) bool
}

func (store *WorkerStore) bind(ctx context.Context,
	scope tenantScope,
) (*workerBound, error) {
	if store == nil || store.bindOwned == nil || !importWorkerScopeUsable(ctx, scope) {
		return nil, app.ErrAuthorizationCheckpoint
	}
	return store.bindOwned(scope)
}
