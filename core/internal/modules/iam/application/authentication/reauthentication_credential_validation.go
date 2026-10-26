package authentication

import "github.com/google/uuid"

func (credential ReauthenticationCredential) baseInvalid(
	lookup ReauthenticationLookup,
) bool {
	portalInvalid := credential.AccessMode == AccessModeOperator &&
		(credential.PortalBindingID != uuid.Nil || credential.PortalBindingRevision != 0) ||
		credential.AccessMode == AccessModePortal &&
			(credential.PortalBindingID != lookup.PortalBindingID ||
				credential.PortalBindingState != "active" || credential.PortalBindingRevision <= 0)
	return credential.UserRevision <= 0 || credential.TokenVersion != lookup.ExpectedTokenVersion ||
		credential.AuthorizationVersion != lookup.ExpectedAuthorizationVersion ||
		credential.PlatformAuthorizationClock < 0 ||
		credential.SecurityEpoch == [32]byte{} ||
		credential.UserStatus != "active" || credential.IdentityKind != "human" ||
		credential.SessionRevision <= 0 || credential.SessionState != "active" ||
		credential.SessionAbsoluteExpiresAt.IsZero() || credential.AccessModeID == uuid.Nil ||
		credential.AccessMode != lookup.AccessMode || credential.AccessModeState != "active" ||
		credential.AccessModeRevision <= 0 || portalInvalid ||
		credential.FamilyID == uuid.Nil || credential.TokenID == uuid.Nil ||
		credential.FamilyRevision <= 0 || credential.TokenRevision <= 0 ||
		credential.FamilyState != "active" || credential.TokenState != "active" ||
		credential.FamilyAbsoluteExpiresAt.IsZero()
}

func (credential ReauthenticationCredential) proofInvalid() bool {
	switch credential.Method {
	case ReauthenticationPassword:
		return credential.AuthenticatorID != uuid.Nil ||
			credential.PasswordCredentialVersion <= 0 ||
			credential.PasswordCredentialRevision <= 0 ||
			credential.Password.Algorithm != AlgorithmArgon2idV1 ||
			credential.Password.invalid()
	case ReauthenticationWebAuthn:
		return credential.AuthenticatorID == uuid.Nil ||
			credential.AuthenticatorKind != "webauthn_v1" ||
			credential.AuthenticatorState != "active" ||
			credential.AuthenticatorRevision <= 0 || !credential.WebAuthn.valid()
	case ReauthenticationTOTP:
		return credential.AuthenticatorID == uuid.Nil ||
			credential.AuthenticatorKind != "totp_v1" ||
			credential.AuthenticatorState != "active" ||
			credential.AuthenticatorRevision <= 0 || !credential.TOTP.valid()
	default:
		return true
	}
}
