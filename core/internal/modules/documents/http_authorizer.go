package documents

import (
	"context"
	"errors"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type applicationAuthorizer struct{}

var errApplicationForbidden = errors.New("Documents application authorization failed")

func (applicationAuthorizer) Authorize(ctx context.Context,
	scope platformtx.TenantScope, permission string,
) error {
	evidence, found := foundation.AuthorityEvidenceFromContext(ctx)
	if !found || evidence.Permission != permission ||
		evidence.Operation != scope.OperationID() ||
		evidence.ActorID.String() != scope.ActorID().String() ||
		evidence.TenantID.String() != scope.TenantID().String() ||
		platformtx.NetworkAllowed(ctx) ||
		platformtx.RequireOwner(scope, "documents") != nil {
		return errApplicationForbidden
	}
	return nil
}

type archiveAuthorizer struct{ application applicationAuthorizer }

func (authorizer archiveAuthorizer) Authorize(ctx context.Context,
	scope platformtx.TenantScope, permission string,
) (uint64, error) {
	if err := authorizer.application.Authorize(ctx, scope, permission); err != nil {
		return 0, err
	}
	evidence, _ := foundation.AuthorityEvidenceFromContext(ctx)
	return uint64(evidence.AuthorizationVersion), nil
}

func (archiveAuthorizer) Refresh(ctx context.Context, tenantID, actorID foundation.ID,
	revision uint64, itemIDs []foundation.ID,
) error {
	evidence, found := foundation.AuthorityEvidenceFromContext(ctx)
	if !found || evidence.TenantID != tenantID || evidence.ActorID != actorID ||
		uint64(evidence.AuthorizationVersion) != revision || len(itemIDs) == 0 {
		return errApplicationForbidden
	}
	for _, itemID := range itemIDs {
		if !itemID.UUIDv7() {
			return errApplicationForbidden
		}
	}
	return nil
}

type stepUpVerifier struct{}

func (stepUpVerifier) Verify(ctx context.Context, scope platformtx.TenantScope,
	operation string, target foundation.ID,
) (time.Time, error) {
	evidence, found := foundation.AuthorityEvidenceFromContext(ctx)
	if !found || !target.UUIDv7() || operation != "documents.item.purge" ||
		evidence.Operation != operation || evidence.TenantID.String() != scope.TenantID().String() ||
		evidence.StepUpAt.IsZero() {
		return time.Time{}, errApplicationForbidden
	}
	return evidence.StepUpAt, nil
}
