package postgres

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"

func (row *reauthenticationCredentialRow) decodeWebAuthn() error {
	if row.authenticatorID == nil || row.authenticatorState == nil ||
		row.authenticatorKind == nil || row.authenticatorRevision == nil ||
		row.webRPID == nil || row.webOriginPolicy == nil ||
		row.webSignCount == nil || row.webBackupEligible == nil ||
		row.webBackupState == nil || row.webUV == nil || row.webRevision == nil ||
		len(row.webRPIDSHA) != 32 {
		return errAuthenticationStorage
	}
	credential := &row.credential
	credential.AuthenticatorID = *row.authenticatorID
	credential.AuthenticatorState = *row.authenticatorState
	credential.AuthenticatorKind = *row.authenticatorKind
	credential.AuthenticatorRevision = *row.authenticatorRevision
	credential.WebAuthn = authentication.WebAuthnCredential{
		CredentialID:             append([]byte(nil), row.webCredentialID...),
		PublicKeyCOSE:            append([]byte(nil), row.webPublicKey...),
		RPID:                     *row.webRPID,
		OriginPolicyVersion:      *row.webOriginPolicy,
		SignCount:                *row.webSignCount,
		BackupEligible:           *row.webBackupEligible,
		BackupState:              *row.webBackupState,
		UserVerificationRequired: *row.webUV,
		Revision:                 *row.webRevision,
	}
	copy(credential.WebAuthn.RPIDSHA256[:], row.webRPIDSHA)
	return nil
}
