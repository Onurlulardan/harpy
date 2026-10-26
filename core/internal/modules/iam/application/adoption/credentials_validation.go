package adoption

import "regexp"

var adoptedBcryptPattern = regexp.MustCompile(`^\$2[aby]\$(0[4-9]|[12][0-9]|3[01])\$[./A-Za-z0-9]{53}$`)

func (state *validationState) validateCredential(user SourceUser) {
	if user.PasswordHash == "" {
		if !user.PortalSetupRequired {
			state.add(issue("credential_missing", "credentials", user.ID.String()))
		}
		return
	}
	if !adoptedBcryptPattern.MatchString(user.PasswordHash) {
		state.add(issue("credential_format_invalid", "credentials", user.ID.String()))
	}
	if user.PortalSetupRequired && user.PortalSetupCompleted {
		state.add(issue("portal_setup_state_invalid", "credentials", user.ID.String()))
	}
}
