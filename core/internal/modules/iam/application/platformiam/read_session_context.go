package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (service *Service) ReadSessionContext(
	ctx context.Context, scope platformtx.TenantScope, query platformdomain.SessionContextQuery,
) (platformdomain.SessionContext, error) {
	if service == nil || query.Validate() != nil || !queryScope(ctx, scope, query.ActorID) {
		return platformdomain.SessionContext{}, platformdomain.ErrInvalidQuery
	}
	result, err := service.store.ReadSessionContext(ctx, scope, query)
	if err != nil {
		return platformdomain.SessionContext{}, err
	}
	if result.Validate() != nil ||
		result.AuthorizationVersion != query.ExpectedAuthorizationVersion {
		return platformdomain.SessionContext{}, platformdomain.ErrStorage
	}
	return result, nil
}
