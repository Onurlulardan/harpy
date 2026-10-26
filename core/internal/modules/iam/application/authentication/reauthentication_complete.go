package authentication

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator Authenticator) completeReauthentication(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	client ClientID,
	credential ReauthenticationCredential,
	proof ReauthenticationProof,
) (Grant, error) {
	successor, err := authenticator.refresh.Generate()
	if err != nil || nilInterface(successor) {
		return Grant{}, ErrAuthenticationCrypto
	}
	destroySuccessor := true
	defer func() {
		if destroySuccessor {
			successor.Destroy()
		}
	}()
	selector, err := successor.Selector()
	if err != nil {
		return Grant{}, ErrAuthenticationCrypto
	}
	keyID, key, err := authenticator.keys.Active()
	if err != nil {
		return Grant{}, ErrAuthenticationCrypto
	}
	secretHMAC, err := successor.SecretHMAC(key[:])
	clear(key[:])
	if err != nil {
		return Grant{}, ErrAuthenticationCrypto
	}
	defer clear(secretHMAC[:])
	identifiers, err := authenticator.nextRotationIdentifiers()
	if err != nil {
		return Grant{}, err
	}
	completed, err := authenticator.port.CompleteReauthentication(
		ctx, scope, credential.completion(
			proof, identifiers.token, selector, keyID, secretHMAC,
		),
	)
	if err != nil || completed.validate(credential) != nil {
		return Grant{}, ErrAuthenticationStorage
	}
	issuedAt := completed.IssuedAt.UTC().Truncate(time.Second)
	access, err := authenticator.signAccess(
		credential.stepUpClaims(scope.TenantID(), client, identifiers.jti, issuedAt),
	)
	if err != nil {
		return Grant{}, err
	}
	grant, err := newGrant(
		access, successor, credential.SessionID, credential.FamilyAbsoluteExpiresAt,
	)
	if err != nil {
		return Grant{}, err
	}
	destroySuccessor = false
	return grant, nil
}
