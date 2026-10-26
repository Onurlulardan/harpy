package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func (store *store) LoginCredential(
	ctx context.Context,
	scope authentication.CredentialScope,
) (authentication.LoginCredential, error) {
	bound, err := store.bindLogin(ctx, scope)
	if err != nil {
		return authentication.LoginCredential{}, err
	}
	mode, ok := loginAccessMode(scope)
	if !ok {
		return authentication.LoginCredential{}, errAuthenticationStorage
	}
	return bound.loginCredential(ctx, scope, mode)
}

func (bound *boundStore) loginCredential(
	ctx context.Context,
	scope authentication.CredentialScope,
	mode authentication.AccessMode,
) (authentication.LoginCredential, error) {
	var result authentication.LoginCredential
	var portalID *uuid.UUID
	var portalRevision *int64
	var salt, tag []byte
	var memory *int32
	var iterations, parallelism *int16
	var legacy *string
	binding := scope.LocatorBindingSHA256()
	err := bound.database.QueryRow(ctx, loginCredentialSQL,
		scope.SourceUserID(), scope.SourceIdentifierID(),
		string(scope.SourceIdentifierKind()), scope.SourceRevision(),
		binding[:], string(mode),
	).Scan(
		&result.IdentifierID, &result.UserID, &result.UserRevision,
		&result.TokenVersion, &result.AuthorizationVersion, &result.AccessModeID,
		&result.AccessModeRevision, &portalID, &portalRevision,
		&result.CredentialVersion, &result.CredentialRevision,
		&result.Password.Algorithm, &salt, &tag, &memory, &iterations,
		&parallelism, &legacy, &result.CredentialCreatedAt,
	)
	if err != nil {
		return authentication.LoginCredential{}, errAuthenticationStorage
	}
	result.AccessMode = mode
	if portalID != nil {
		result.PortalBindingID = *portalID
	}
	if portalRevision != nil {
		result.PortalBindingRevision = *portalRevision
	}
	passwordCredential, err := decodePasswordCredential(
		result.Password.Algorithm, salt, tag, memory, iterations,
		parallelism, legacy, result.CredentialCreatedAt,
	)
	clear(salt)
	clear(tag)
	if err != nil {
		return authentication.LoginCredential{}, err
	}
	result.Password = passwordCredential
	return result, nil
}

func loginAccessMode(
	scope authentication.CredentialScope,
) (authentication.AccessMode, bool) {
	switch string(scope.SourceIdentifierKind()) {
	case "email_v1":
		return authentication.AccessModeOperator, true
	case "portal_v1":
		return authentication.AccessModePortal, true
	default:
		return "", false
	}
}
