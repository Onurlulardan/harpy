package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
)

func (store *store) PasswordCredentialLookup(ctx context.Context,
	scope tenantScope,
	query lifecycle.PasswordChangeCommand) (lifecycle.PasswordCredential, error) {
	if !snapshotUsable(ctx, scope) || query.Evidence.Validate() != nil ||
		scope.ActorID() != query.Evidence.ActorID {
		return lifecycle.PasswordCredential{}, lifecycle.ErrInvalid
	}
	bound, err := store.bindTenant(scope)
	if err != nil {
		return lifecycle.PasswordCredential{}, err
	}
	e := query.Evidence
	var result lifecycle.PasswordCredential
	var legacyBcryptHash *string
	err = bound.database.QueryRow(ctx, passwordCredentialReadSQL, query.UserID,
		e.ActorID, e.PermissionID, e.ExpectedTokenVersion,
		e.ExpectedAuthorizationVersion).Scan(&result.Algorithm, &result.Salt, &result.Tag,
		&result.MemoryKiB, &result.Iterations, &result.Parallelism,
		&legacyBcryptHash, &result.Version, &result.Revision, &result.CreatedAt)
	if err != nil {
		return lifecycle.PasswordCredential{}, mapError(err)
	}
	if legacyBcryptHash != nil {
		result.LegacyBcryptHash = *legacyBcryptHash
	}
	return result, nil
}
