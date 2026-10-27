package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

func (store *store) StartReauthentication(
	ctx context.Context,
	scope authentication.CredentialScope,
	creation authentication.ChallengeCreation,
) (authentication.ChallengeCreated, error) {
	bound, err := store.bindReauthentication(ctx, scope)
	if err != nil {
		return authentication.ChallengeCreated{}, err
	}
	return bound.persistReauthenticationChallenge(ctx, scope, creation)
}

func (bound *boundStore) persistReauthenticationChallenge(
	ctx context.Context,
	scope authentication.CredentialScope,
	creation authentication.ChallengeCreation,
) (authentication.ChallengeCreated, error) {
	portalID, _ := portalArguments(creation.PortalBindingID, 1)
	var authenticatorID any
	if creation.AuthenticatorID != uuid.Nil {
		authenticatorID = creation.AuthenticatorID
	}
	binding := scope.LocatorBindingSHA256()
	challengeHMAC := creation.ChallengeHMAC
	contextSHA := creation.ContextSHA256
	purposeSHA := creation.PurposeSHA256
	originSHA := creation.OriginSHA256
	var result authentication.ChallengeCreated
	var credentialID []byte
	var rpID *string
	var userVerification *bool
	err := bound.database.QueryRow(ctx, createReauthenticationChallengeSQL,
		scope.TenantID(), binding[:], creation.ID, creation.UserID,
		creation.SessionID, creation.ExpectedTokenVersion,
		creation.ExpectedAuthorizationVersion, string(creation.AccessMode),
		portalID, string(creation.Method), authenticatorID,
		creation.HMACKeyID, challengeHMAC[:], contextSHA[:], purposeSHA[:],
		creation.ActionPolicyVersion, originSHA[:],
		creation.RequirePhishingResistant,
	).Scan(
		&result.Revision, &result.IssuedAt, &result.ExpiresAt,
		&credentialID, &rpID, &userVerification,
	)
	if err != nil {
		clear(credentialID)
		return authentication.ChallengeCreated{}, errAuthenticationStorage
	}
	result.CredentialID = append([]byte(nil), credentialID...)
	clear(credentialID)
	if rpID != nil {
		result.RPID = *rpID
	}
	if userVerification != nil {
		result.UserVerificationRequired = *userVerification
	}
	return result, nil
}
