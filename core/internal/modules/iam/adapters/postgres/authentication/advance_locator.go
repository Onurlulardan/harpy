package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (bound *boundStore) advanceLocator(
	ctx context.Context,
	scope authentication.CredentialScope,
	projection locatorAdvance,
) error {
	if !projection.valid() {
		return errAuthenticationStorage
	}
	_, err := bound.database.Execute(ctx, advanceRefreshLocatorSQL,
		projection.selector, scope.TenantID(), scope.AuthorityGroupGeneration(),
		scope.IAMGeneration(), *projection.expected, *projection.revision,
		projection.sourceSHA256, *projection.state, *projection.projectedAt,
		*projection.retainUntil,
	)
	if err != nil {
		return errAuthenticationStorage
	}
	return nil
}
