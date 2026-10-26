package iam

import (
	"context"

	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) finishIdentityMutation(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityfacade.MutationEvidence,
	aggregateID uuid.UUID,
	resourceKind string,
	commandKey string,
	replayed bool,
) error {
	if replayed {
		return nil
	}
	if application == nil || nilIdentityDependency(application.changes) {
		return changeapp.ErrSink
	}
	change := changeapp.NewChange(
		evidence, identityfacade.IdentityID(aggregateID), resourceKind, commandKey,
	)
	if change.Validate() != nil ||
		application.changes.WriteChange(ctx, scope, change) != nil {
		return changeapp.ErrSink
	}
	return nil
}
