package lifecycle

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (service *Service) PasswordCredentialLookup(
	ctx context.Context,
	scope platformtx.TenantScope,
	query PasswordChangeCommand,
) (PasswordCredential, error) {
	if !service.Valid() || !tenantSnapshotUsable(ctx, scope) ||
		query.Evidence.Validate() != nil || query.UserID == uuid.Nil {
		return PasswordCredential{}, ErrInvalid
	}
	return service.store.PasswordCredentialLookup(ctx, scope, query)
}
