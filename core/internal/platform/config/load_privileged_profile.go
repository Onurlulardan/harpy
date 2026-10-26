package config

func loadPrivilegedProfile(
	values environmentValues,
	dsnKey string,
	expectedPrincipal string,
	receiptKey string,
) (Environment, PrivilegedPostgres, ReleaseAttestation, LogLevel, error) {
	rawEnvironment, found := values.lookup("HARPY_ENVIRONMENT")
	if !found {
		return "", PrivilegedPostgres{}, ReleaseAttestation{}, "",
			configError("HARPY_ENVIRONMENT", "is required")
	}
	environment, err := parseEnvironment(rawEnvironment)
	if err != nil {
		return "", PrivilegedPostgres{}, ReleaseAttestation{}, "", err
	}
	postgres, err := loadPrivilegedPostgres(values, environment, dsnKey, expectedPrincipal)
	if err != nil {
		return "", PrivilegedPostgres{}, ReleaseAttestation{}, "", err
	}
	release, err := loadReleaseAttestation(values, receiptKey)
	if err != nil {
		return "", PrivilegedPostgres{}, ReleaseAttestation{}, "", err
	}
	logLevel, err := parseLogLevel(values.valueOr("HARPY_LOG_LEVEL", "info"))
	if err != nil {
		return "", PrivilegedPostgres{}, ReleaseAttestation{}, "", err
	}
	return environment, postgres, release, logLevel, nil
}
