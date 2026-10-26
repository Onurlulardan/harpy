package authentication

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (authenticator Authenticator) Issue(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	request IssueRequest,
) (Grant, error) {
	if request.Secret != nil {
		defer request.Secret.Destroy()
	}
	if !authenticator.valid || request.validate() != nil ||
		!validCredentialScope(
			ctx, scope, platformtx.CredentialLogin,
			platformtx.CredentialLoginDirectory,
		) {
		return Grant{}, ErrAuthenticationContract
	}
	credential, err := authenticator.port.LoginCredential(ctx, scope)
	if err != nil || credential.validate(scope) != nil ||
		!clientMatchesMode(request.Client, credential.AccessMode) {
		return Grant{}, ErrInvalidCredentials
	}
	now := authenticator.clock().UTC()
	if now.IsZero() {
		return Grant{}, ErrAuthenticationCrypto
	}
	rehash, err := authenticator.verifyPassword(request.Secret, credential.Password, now)
	if err != nil {
		return Grant{}, err
	}
	used, err := authenticator.port.UsePassword(ctx, scope, PasswordUse{
		UserID:                       credential.UserID,
		ExpectedUserRevision:         credential.UserRevision,
		ExpectedTokenVersion:         credential.TokenVersion,
		ExpectedAuthorizationVersion: credential.AuthorizationVersion,
		ExpectedCredentialVersion:    credential.CredentialVersion,
		ExpectedCredentialRevision:   credential.CredentialRevision,
		Rehash:                       rehash,
	})
	if err != nil || used.validate() != nil {
		return Grant{}, ErrAuthenticationStorage
	}
	refresh, err := authenticator.refresh.Generate()
	if err != nil || refresh == nil {
		return Grant{}, ErrAuthenticationCrypto
	}
	destroyRefresh := true
	defer func() {
		if destroyRefresh {
			refresh.Destroy()
		}
	}()
	selector, err := refresh.Selector()
	if err != nil {
		return Grant{}, ErrAuthenticationCrypto
	}
	hashKeyID, hashKey, err := authenticator.keys.Active()
	if err != nil {
		return Grant{}, ErrAuthenticationCrypto
	}
	secretHMAC, err := refresh.SecretHMAC(hashKey[:])
	clear(hashKey[:])
	if err != nil {
		return Grant{}, ErrAuthenticationCrypto
	}
	identifiers, err := authenticator.nextIdentifierSet()
	if err != nil {
		return Grant{}, err
	}
	issued, err := authenticator.port.Issue(ctx, scope, SessionIssue{
		UserID: credential.UserID, ExpectedUserRevision: used.UserRevision,
		ExpectedTokenVersion:         used.TokenVersion,
		ExpectedAuthorizationVersion: used.AuthorizationVersion,
		AccessModeID:                 credential.AccessModeID, AccessMode: credential.AccessMode,
		ExpectedAccessModeRevision:    credential.AccessModeRevision,
		PortalBindingID:               credential.PortalBindingID,
		ExpectedPortalBindingRevision: credential.PortalBindingRevision,
		SessionID:                     identifiers.session, FamilyID: identifiers.family,
		TokenID: identifiers.token, Device: request.Device, Selector: selector,
		HashKeyID: hashKeyID, SecretHMAC: secretHMAC,
	})
	clear(secretHMAC[:])
	if err != nil || issued.validate() != nil {
		return Grant{}, ErrAuthenticationStorage
	}
	issuedAt := issued.IssuedAt.UTC().Truncate(time.Second)
	claimsInput := AccessClaims{
		Subject: credential.UserID, Session: identifiers.session,
		Tenant: scope.TenantID(), Client: request.Client, IssuedAt: issuedAt,
		JTI: identifiers.jti, TokenVersion: used.TokenVersion,
		AuthorizationVersion:       used.AuthorizationVersion,
		PlatformAuthorizationClock: issued.PlatformAuthorizationClock,
		SecurityEpoch:              issued.SecurityEpoch,
		AccessMode:                 credential.AccessMode, AuthenticationTime: issuedAt,
		Assurance: AssuranceBaseline,
		Methods:   []AuthMethod{AuthMethodPassword},
	}
	if credential.PortalBindingID != uuid.Nil {
		binding := credential.PortalBindingID
		claimsInput.PortalBinding = &binding
	}
	access, err := authenticator.signAccess(claimsInput)
	if err != nil {
		return Grant{}, err
	}
	grant, err := newGrant(access, refresh, identifiers.session, issued.AbsoluteExpiresAt)
	if err != nil {
		return Grant{}, err
	}
	destroyRefresh = false
	return grant, nil
}
