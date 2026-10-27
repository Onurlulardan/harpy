package transaction

type CredentialPurpose string

const (
	CredentialLogin             CredentialPurpose = "login"
	CredentialRefresh           CredentialPurpose = "refresh"
	CredentialPasswordReset     CredentialPurpose = "password_reset"
	CredentialInitialAdminSetup CredentialPurpose = "initial_admin_setup"
	CredentialInitialPassword   CredentialPurpose = "initial_password"
	CredentialEmailVerification CredentialPurpose = "email_verification"
	CredentialPortalSetup       CredentialPurpose = "portal_setup"
	CredentialInvitation        CredentialPurpose = "invitation"
	CredentialReauthentication  CredentialPurpose = "reauth"
)

func (purpose CredentialPurpose) valid() bool {
	switch purpose {
	case CredentialLogin, CredentialRefresh, CredentialPasswordReset,
		CredentialInitialAdminSetup,
		CredentialInitialPassword,
		CredentialEmailVerification, CredentialPortalSetup,
		CredentialInvitation, CredentialReauthentication:
		return true
	default:
		return false
	}
}
