package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func (bound *boundStore) createLocator(
	ctx context.Context,
	scope authentication.CredentialScope,
	tokenID uuid.UUID,
	familyID uuid.UUID,
	projection locatorCreate,
) error {
	if tokenID == uuid.Nil || familyID == uuid.Nil || !projection.valid() {
		return errAuthenticationStorage
	}
	_, err := bound.database.Execute(ctx, createRefreshLocatorSQL,
		projection.selector, scope.TenantID(), tokenID, familyID,
		scope.AuthorityGroupGeneration(), scope.IAMGeneration(),
		*projection.revision, projection.sourceSHA256, *projection.state,
		*projection.issuedAt, *projection.projectedAt,
		*projection.familyExpires, *projection.retainUntil,
	)
	if err != nil {
		return errAuthenticationStorage
	}
	return nil
}
