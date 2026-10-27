package postgres

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"

func (row *reauthenticationCredentialRow) decodeTOTP() error {
	if row.authenticatorID == nil || row.authenticatorState == nil ||
		row.authenticatorKind == nil || row.authenticatorRevision == nil ||
		row.totpProfile == nil || row.totpCipherSuite == nil ||
		row.totpKMSKeyID == nil || row.totpRevision == nil ||
		len(row.totpAAD) != 32 {
		return errAuthenticationStorage
	}
	credential := &row.credential
	credential.AuthenticatorID = *row.authenticatorID
	credential.AuthenticatorState = *row.authenticatorState
	credential.AuthenticatorKind = *row.authenticatorKind
	credential.AuthenticatorRevision = *row.authenticatorRevision
	credential.TOTP = authentication.TOTPEnvelope{
		Profile:              *row.totpProfile,
		CipherSuite:          *row.totpCipherSuite,
		KMSKeyID:             *row.totpKMSKeyID,
		Nonce:                append([]byte(nil), row.totpNonce...),
		Ciphertext:           append([]byte(nil), row.totpCiphertext...),
		AuthenticationTag:    append([]byte(nil), row.totpTag...),
		LastAcceptedTimeStep: row.totpLastStep,
		Revision:             *row.totpRevision,
	}
	copy(credential.TOTP.AADSHA256[:], row.totpAAD)
	return nil
}
