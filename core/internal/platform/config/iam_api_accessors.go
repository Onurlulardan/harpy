package config

func (config IAMAPI) HTTPProfile() (IAMHTTPProfile, error) {
	if err := config.Validate(); err != nil {
		return IAMHTTPProfile{}, err
	}
	return config.http, nil
}

func (config IAMAPI) AccessTokenKeyring() (IAMAccessTokenKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMAccessTokenKeyring{}, err
	}
	return config.access, nil
}

func (config IAMAPI) RefreshTokenKeyring() (IAMRefreshKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMRefreshKeyring{}, err
	}
	return config.refresh, nil
}

func (config IAMAPI) CSRFKeyring() (IAMCSRFKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMCSRFKeyring{}, err
	}
	return config.csrf, nil
}

func (config IAMAPI) LoginDirectoryKeyring() (IAMDirectoryKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMDirectoryKeyring{}, err
	}
	return config.directory, nil
}

func (config IAMAPI) AuthThrottleKeyring() (IAMThrottleKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMThrottleKeyring{}, err
	}
	return config.throttle, nil
}

func (config IAMAPI) SessionNetworkKeyring() (IAMSessionNetworkKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMSessionNetworkKeyring{}, err
	}
	return config.sessionNetwork, nil
}

func (config IAMAPI) OneTimeTokenKeyring() (IAMOneTimeKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMOneTimeKeyring{}, err
	}
	return config.oneTime, nil
}

func (config IAMAPI) RecoveryCodeKeyring() (IAMRecoveryKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMRecoveryKeyring{}, err
	}
	return config.recovery, nil
}

func (config IAMAPI) PasswordBlocklists() (IAMPasswordBlocklists, error) {
	if err := config.Validate(); err != nil {
		return IAMPasswordBlocklists{}, err
	}
	return config.passwords, nil
}

func (config IAMAPI) ActorBridge() (IAMActorBridge, error) {
	if err := config.Validate(); err != nil {
		return IAMActorBridge{}, err
	}
	return config.actorBridge, nil
}
