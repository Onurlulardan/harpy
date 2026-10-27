package main

func sharedRuntimeSecretNames() []string {
	return []string{
		"catalog-package-public-keyring.json",
		"documents-kek-keyring.json",
		"analytics-dashboard-export-fcart-kek-keyring.json",
		"gcp-subject-token.jwt",
		"gcp-wif.json",
		"iam-login-directory-keyring.json",
		"iam-one-time-token-keyring.json",
		"iam-system-export-fcart-kek-keyring.json",
		"iam-system-import-source-fcart-kek-keyring.json",
		"local-ca.crt",
		"postgres-runtime-dsn.txt",
		"rest-query-cursor-keyring.json",
		"s3-access-key-id.txt",
		"s3-secret-access-key.txt",
	}
}

func apiSecretNames() []string {
	names := append([]string{}, sharedRuntimeSecretNames()...)
	return append(names,
		"actor-bridge-client.crt",
		"actor-bridge-client.key",
		"actor-bridge-public-keys.json",
		"actor-bridge-server-ca.crt",
		"actor-bridge-signing-seed",
		"actor-bridge-targets.json",
		"iam-access-token-keyring.json",
		"iam-auth-throttle-keyring.json",
		"iam-csrf-keyring.json",
		"iam-password-blocklists.json",
		"iam-password-common.bin",
		"iam-password-compromised.bin",
		"iam-recovery-code-keyring.json",
		"iam-refresh-token-keyring.json",
		"iam-session-network-keyring.json",
		"notifications-consumer-keyring.json",
	)
}

func workerSecretNames() []string {
	names := append([]string{}, sharedRuntimeSecretNames()...)
	return append(names,
		"invitation-template.json",
		"local-authority-ack-signing-seed",
		"notifications-kek-keyring.json",
		"rabbitmq-url.txt",
		"smtp-password.txt",
	)
}
