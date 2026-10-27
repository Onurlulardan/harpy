package application

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type ContextAuthorizer struct{}

func (ContextAuthorizer) Authorize(
	ctx context.Context, scope platformtx.TenantScope, permission Permission,
) error {
	evidence, ok := FromContext(ctx)
	if !ok || !evidence.Authorizes(scope, string(permission)) {
		return ErrForbidden
	}
	return nil
}

var _ Authorizer = ContextAuthorizer{}
