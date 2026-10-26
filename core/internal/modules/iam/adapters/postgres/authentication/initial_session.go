package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) InitialSession(
	ctx context.Context,
	scope authentication.CredentialScope,
	issue authentication.SessionIssue,
) (authentication.SessionIssued, error) {
	bound, err := store.bindInitialRoute(ctx, scope)
	if err != nil {
		return authentication.SessionIssued{}, err
	}
	return bound.issue(ctx, scope, issue)
}
