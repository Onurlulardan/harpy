package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func (store *store) InitialPasswordCredential(
	ctx context.Context,
	scope authentication.CredentialScope,
) (authentication.InitialPasswordCredential, error) {
	bound, err := store.bindInitialLogin(ctx, scope)
	if err != nil {
		return authentication.InitialPasswordCredential{}, err
	}
	var result authentication.InitialPasswordCredential
	var salt, tag []byte
	var memory int32
	var iterations, parallelism int16
	binding := scope.LocatorBindingSHA256()
	err = bound.database.QueryRow(ctx, initialPasswordCredentialSQL,
		scope.SourceUserID(), scope.SourceIdentifierID(),
		string(scope.SourceIdentifierKind()), scope.SourceRevision(),
		binding[:],
	).Scan(&result.IdentifierID, &result.UserID, &result.UserRevision,
		&result.TokenVersion, &result.AuthorizationVersion, &result.AccessModeID,
		&result.AccessModeRevision, &result.InstallationOperationID,
		&result.CredentialID, &result.CredentialVersion, &result.CredentialRevision,
		&result.Password.Algorithm, &salt, &tag, &memory, &iterations,
		&parallelism, &result.CreatedAt)
	if err != nil {
		clear(salt)
		clear(tag)
		return authentication.InitialPasswordCredential{}, errAuthenticationStorage
	}
	credential, err := decodePasswordCredential(result.Password.Algorithm,
		salt, tag, &memory, &iterations, &parallelism, nil, result.CreatedAt)
	clear(salt)
	clear(tag)
	if err != nil {
		return authentication.InitialPasswordCredential{}, err
	}
	result.Password = credential
	return result, nil
}
