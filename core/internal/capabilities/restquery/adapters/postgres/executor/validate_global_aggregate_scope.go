package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func validateGlobalAggregateScope(
	ctx context.Context,
	scope transaction.GlobalScope,
	batch compiler.AggregateBatch,
	plan resolved.AuthorizedPlan,
) error {
	if ctx == nil || ctx.Err() != nil || transaction.NetworkAllowed(ctx) ||
		!transaction.ContextOwnsGlobalScope(ctx, scope) ||
		transaction.RequireActiveGlobal(scope, transaction.SnapshotQuery) != nil {
		return ErrContract
	}
	binding, err := plan.SnapshotBinding()
	if err != nil {
		return ErrContract
	}
	kind, kindErr := binding.Scope()
	environmentID, environmentErr := binding.ScopeID()
	subjectID, hasSubject, subjectErr := binding.SubjectID()
	audience, audienceErr := binding.Audience()
	operationID, operationErr := binding.OperationID()
	if kindErr != nil || environmentErr != nil || subjectErr != nil ||
		audienceErr != nil || operationErr != nil ||
		kind != resolved.SnapshotScopeGlobal || environmentID != scope.EnvironmentID() ||
		operationID != scope.OperationID() {
		return ErrContract
	}
	if audience == resolved.SnapshotAudienceGlobalPublic {
		if hasSubject || subjectID != uuid.Nil || scope.ActorID() != uuid.Nil {
			return ErrContract
		}
	} else if audience != resolved.SnapshotAudienceGlobalActor || !hasSubject ||
		subjectID == uuid.Nil || subjectID != scope.ActorID() {
		return ErrContract
	}
	catalog, err := batch.Catalog()
	owner, ownerErr := catalog.Owner()
	if err != nil || ownerErr != nil || transaction.RequireGlobalOwner(scope, owner) != nil {
		return ErrContract
	}
	return nil
}
