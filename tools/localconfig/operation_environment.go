package main

const releaseRoot = "/run/release/current/"

func initializeEnvironment() profileEnvironment {
	return profileEnvironment{
		"HARPY_API_PASSWORD_FILE":         secretRoot + "postgres-api-password.txt",
		"HARPY_FOUNDATION_ADMIN_DSN_FILE": secretRoot + "postgres-bootstrap-dsn.txt",
		"HARPY_FOUNDATION_DATABASE":       "harpy",
		"HARPY_GOOSE_PASSWORD_FILE":       secretRoot + "postgres-goose-password.txt",
		"HARPY_WORKER_PASSWORD_FILE":      secretRoot + "postgres-worker-password.txt",
	}
}

func operatorEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(secretRoot + "postgres-runtime-dsn.txt")
	values["HARPY_IAM_LOGIN_DIRECTORY_KEYRING_FILE"] =
		secretRoot + "iam-login-directory-keyring.json"
	values["HARPY_INSTALLATION_MANIFEST_PATH"] = "/run/installation/manifest.json"
	values["HARPY_INITIAL_ADMIN_PASSWORD_FILE"] =
		"/run/installation/initial-admin-password"
	values["HARPY_RECOVERY_SIGNING_DIRECTORY"] = "/run/recovery-signing/current"
	return values
}
