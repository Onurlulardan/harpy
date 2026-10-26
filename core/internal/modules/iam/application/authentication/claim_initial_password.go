package authentication

import (
	"context"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (authenticator InitialPasswordAuthenticator) Claim(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	temporary PasswordSecret,
) (InitialPasswordChallenge, error) {
	if temporary != nil {
		defer temporary.Destroy()
	}
	if !authenticator.valid || nilInterface(temporary) || !validCredentialScope(
		ctx, scope, platformtx.CredentialInitialPassword,
		platformtx.CredentialLoginDirectory,
	) {
		return InitialPasswordChallenge{}, ErrAuthenticationContract
	}
	credential, err := authenticator.port.InitialPasswordCredential(ctx, scope)
	if err != nil || credential.validate(scope) != nil {
		return InitialPasswordChallenge{}, ErrInvalidCredentials
	}
	matched, err := authenticator.base.password.VerifyArgon2id(
		temporary, credential.Password.Argon2id,
	)
	if err != nil {
		return InitialPasswordChallenge{}, ErrAuthenticationCrypto
	}
	if !matched {
		return InitialPasswordChallenge{}, ErrInvalidCredentials
	}
	material, err := generateInitialPasswordMaterial(authenticator.randomness)
	if err != nil {
		return InitialPasswordChallenge{}, err
	}
	destroyMaterial := true
	defer func() {
		if destroyMaterial {
			material.Destroy()
		}
	}()
	claim, err := authenticator.claimCommand(scope, credential, material)
	if err != nil {
		return InitialPasswordChallenge{}, err
	}
	claimed, err := authenticator.port.ClaimInitialPassword(ctx, scope, claim)
	if err != nil || claimed.Revision != 1 || claimed.IssuedAt.IsZero() ||
		!claimed.ExpiresAt.Equal(claimed.IssuedAt.Add(InitialPasswordLifetime)) ||
		claimed.IssuedAt.Location() != time.UTC || claimed.ExpiresAt.Location() != time.UTC {
		return InitialPasswordChallenge{}, ErrAuthenticationStorage
	}
	challenge, err := newInitialPasswordChallenge(material, claimed.ExpiresAt)
	if err != nil {
		return InitialPasswordChallenge{}, err
	}
	destroyMaterial = false
	return challenge, nil
}

func (authenticator InitialPasswordAuthenticator) claimCommand(
	scope platformtx.CredentialTenantScope,
	credential InitialPasswordCredential,
	material *InitialPasswordMaterial,
) (InitialPasswordClaim, error) {
	secretID, secretKey, err := authenticator.continuation.Active()
	if err != nil || !validHMACKeyID(secretID) || secretKey == [32]byte{} {
		clear(secretKey[:])
		return InitialPasswordClaim{}, ErrAuthenticationCrypto
	}
	csrfID, csrfKey, err := authenticator.csrf.Active()
	if err != nil || !validHMACKeyID(csrfID) || csrfKey == [32]byte{} || csrfID == secretID {
		clear(secretKey[:])
		clear(csrfKey[:])
		return InitialPasswordClaim{}, ErrAuthenticationCrypto
	}
	binding := credential.binding(scope.TenantID())
	secretHMAC := initialSecretHMAC(secretKey, binding.tenantID, binding.userID,
		binding.operationID, binding.credentialID, binding.credentialVersion,
		material.selector, material.secret)
	csrfHMAC := initialCSRFHMAC(csrfKey, material.selector, material.csrf)
	clear(secretKey[:])
	clear(csrfKey[:])
	return InitialPasswordClaim{InitialPasswordCredential: credential,
		Selector: material.selector, SecretKeyID: secretID, CSRFKeyID: csrfID,
		SecretHMAC: secretHMAC, CSRFHMAC: csrfHMAC}, nil
}
