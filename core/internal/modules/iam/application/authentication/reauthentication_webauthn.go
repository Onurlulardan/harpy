package authentication

func (authenticator Authenticator) verifyWebAuthn(
	request ReauthenticateRequest,
	credential ReauthenticationCredential,
) (ReauthenticationProof, bool, error) {
	challenge, err := request.Challenge.Wire()
	if err != nil {
		return ReauthenticationProof{}, false, ErrAuthenticationCrypto
	}
	defer clear(challenge)
	result, err := authenticator.security.WebAuthn.Verify(
		request.WebAuthn,
		WebAuthnExpectation{
			Challenge:     challenge,
			CredentialID:  credential.WebAuthn.CredentialID,
			PublicKeyCOSE: credential.WebAuthn.PublicKeyCOSE,
			RPID:          credential.WebAuthn.RPID, Origin: request.Origin,
			UserHandle:               credential.UserID[:],
			UserVerificationRequired: credential.WebAuthn.UserVerificationRequired,
		},
	)
	if err != nil {
		return ReauthenticationProof{}, false, err
	}
	quarantine := result.BackupEligible != credential.WebAuthn.BackupEligible ||
		result.BackupState && !result.BackupEligible ||
		credential.WebAuthn.SignCount > 0 && result.SignCount > 0 &&
			int64(result.SignCount) <= credential.WebAuthn.SignCount &&
			!(credential.WebAuthn.BackupEligible && result.BackupState)
	if quarantine {
		return ReauthenticationProof{}, true, ErrInvalidCredentials
	}
	return ReauthenticationProof{
		Method: credential.Method, WebAuthnSignCount: int64(result.SignCount),
		WebAuthnBackupEligible: result.BackupEligible,
		WebAuthnBackupState:    result.BackupState,
	}, false, nil
}
