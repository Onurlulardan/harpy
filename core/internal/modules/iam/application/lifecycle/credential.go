package lifecycle

import lifecyclecredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/lifecyclecredential"

const (
	PurposeEmailVerification CredentialPurpose = "email_verification_v1"
	PurposePasswordReset     CredentialPurpose = "password_reset_v1"
	PurposePortalSetup       CredentialPurpose = "portal_setup_v1"
	PurposeInitialAdminSetup CredentialPurpose = "initial_admin_setup.v1"
)

const credentialWireBytes = lifecyclecredential.WireBytes

type CredentialPurpose string
type OneTimeKeyring = lifecyclecredential.OneTimeKeyring

func (purpose CredentialPurpose) Valid() bool {
	return lifecyclecredential.Purpose(purpose).Valid()
}

func credentialHMAC(key [32]byte, tenantID [16]byte, purpose CredentialPurpose,
	selector [16]byte, secret [32]byte) ([32]byte, error) {
	return lifecyclecredential.Digest(
		key, tenantID, lifecyclecredential.Purpose(purpose), selector, secret,
	)
}

func CredentialBindingSHA256(tenantID [16]byte, purpose CredentialPurpose,
	selector [16]byte) ([32]byte, error) {
	return lifecyclecredential.BindingSHA256(
		tenantID, lifecyclecredential.Purpose(purpose), selector,
	)
}
