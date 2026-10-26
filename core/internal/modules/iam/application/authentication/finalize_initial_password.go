package authentication

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator InitialPasswordAuthenticator) Finalize(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	verified *VerifiedInitialPassword,
	newPassword PasswordSecret,
	device DeviceEvidence,
) (Grant, error) {
	if newPassword != nil {
		defer newPassword.Destroy()
	}
	if !authenticator.valid || nilInterface(newPassword) || device.validate() != nil ||
		!verifiedInitialUsable(ctx, scope, verified) {
		return Grant{}, ErrAuthenticationContract
	}
	defer verified.destroy()
	credential, err := authenticator.base.password.Hash(newPassword)
	if err != nil || nilInterface(credential) || !credential.Valid() {
		return Grant{}, ErrAuthenticationCrypto
	}
	refresh, issue, identifiers, err := authenticator.prepareInitialSession(device)
	if err != nil {
		return Grant{}, err
	}
	defer clear(issue.SecretHMAC[:])
	destroyRefresh := true
	defer func() {
		if destroyRefresh {
			refresh.Destroy()
		}
	}()
	continuation := verified.continuation
	command := InitialPasswordFinalize{
		Selector: continuation.Selector, SecretHMAC: verified.secretHMAC,
		CSRFHMAC: verified.csrfHMAC, ExpectedContinuationRevision: continuation.Revision,
		UserID: continuation.UserID, CredentialID: continuation.CredentialID,
		ExpectedCredentialVersion:  continuation.CredentialVersion,
		ExpectedCredentialRevision: continuation.CredentialRevision,
		Salt:                       credential.Salt(), Tag: credential.Tag(),
		RequestSHA256: initialCompletionRequestSHA256(continuation),
	}
	finalized, err := authenticator.port.FinalizeInitialPassword(ctx, scope, command)
	if err != nil || !finalized.valid(command) {
		return Grant{}, ErrAuthenticationStorage
	}
	issue.UserID = finalized.UserID
	issue.ExpectedUserRevision = finalized.UserRevision
	issue.ExpectedTokenVersion = finalized.TokenVersion
	issue.ExpectedAuthorizationVersion = finalized.AuthorizationVersion
	issue.AccessModeID = finalized.AccessModeID
	issue.AccessMode = AccessModeOperator
	issue.ExpectedAccessModeRevision = finalized.AccessModeRevision
	issued, err := authenticator.port.InitialSession(ctx, scope, issue)
	if err != nil || issued.validate() != nil {
		return Grant{}, ErrAuthenticationStorage
	}
	issuedAt := issued.IssuedAt.UTC().Truncate(time.Second)
	access, err := authenticator.base.signAccess(AccessClaims{
		Subject: finalized.UserID, Session: identifiers.session,
		Tenant: scope.TenantID(), Client: ClientOperatorConsole, IssuedAt: issuedAt,
		JTI: identifiers.jti, TokenVersion: finalized.TokenVersion,
		AuthorizationVersion:       finalized.AuthorizationVersion,
		PlatformAuthorizationClock: issued.PlatformAuthorizationClock,
		SecurityEpoch:              issued.SecurityEpoch, AccessMode: AccessModeOperator,
		AuthenticationTime: issuedAt, Assurance: AssuranceBaseline,
		Methods: []AuthMethod{AuthMethodPassword},
	})
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

func (authenticator InitialPasswordAuthenticator) prepareInitialSession(
	device DeviceEvidence,
) (IssuedRefresh, SessionIssue, identifierSet, error) {
	refresh, err := authenticator.base.refresh.Generate()
	if err != nil || nilInterface(refresh) {
		return nil, SessionIssue{}, identifierSet{}, ErrAuthenticationCrypto
	}
	selector, err := refresh.Selector()
	if err != nil {
		refresh.Destroy()
		return nil, SessionIssue{}, identifierSet{}, ErrAuthenticationCrypto
	}
	keyID, key, err := authenticator.base.keys.Active()
	if err != nil || !validHMACKeyID(keyID) {
		refresh.Destroy()
		clear(key[:])
		return nil, SessionIssue{}, identifierSet{}, ErrAuthenticationCrypto
	}
	secretHMAC, err := refresh.SecretHMAC(key[:])
	clear(key[:])
	if err != nil {
		refresh.Destroy()
		return nil, SessionIssue{}, identifierSet{}, ErrAuthenticationCrypto
	}
	identifiers, err := authenticator.base.nextIdentifierSet()
	if err != nil {
		refresh.Destroy()
		clear(secretHMAC[:])
		return nil, SessionIssue{}, identifierSet{}, err
	}
	issue := SessionIssue{SessionID: identifiers.session, FamilyID: identifiers.family,
		TokenID: identifiers.token, Device: device, Selector: selector,
		HashKeyID: keyID, SecretHMAC: secretHMAC}
	return refresh, issue, identifiers, nil
}
