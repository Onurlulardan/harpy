package recoveryanchor

import (
	"context"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) LoadRegistration(
	ctx context.Context,
	scope platformtx.TenantScope,
	subject anchordomain.Subject,
	operationID string,
) (anchordomain.RegistrationLookup, error) {
	if service == nil || !scopeValid(scope, subject) || operationID == "" {
		return anchordomain.RegistrationLookup{}, anchordomain.ErrInvalid
	}
	return service.store.LoadRegistration(ctx, scope, subject, operationID)
}
