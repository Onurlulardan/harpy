package transaction

type CredentialLocatorKind string

const (
	CredentialLoginDirectory          CredentialLocatorKind = "login_directory_v1"
	CredentialRefreshLocator          CredentialLocatorKind = "refresh_locator_v1"
	CredentialTenantRoute             CredentialLocatorKind = "tenant_route_v1"
	CredentialSessionLocator CredentialLocatorKind = "session_locator_v1"
)

func (kind CredentialLocatorKind) validFor(purpose CredentialPurpose) bool {
	switch purpose {
	case CredentialLogin:
		return kind == CredentialLoginDirectory
	case CredentialPasswordReset, CredentialInitialPassword,
		CredentialInitialAdminSetup:
		return kind == CredentialLoginDirectory || kind == CredentialTenantRoute
	case CredentialRefresh:
		return kind == CredentialRefreshLocator
	case CredentialReauthentication:
		return kind == CredentialSessionLocator
	case CredentialEmailVerification,
		CredentialPortalSetup, CredentialInvitation:
		return kind == CredentialTenantRoute
	default:
		return false
	}
}
