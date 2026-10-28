package main

func workerEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(set, secretRoot+"postgres-runtime-dsn.txt")
	values["GOMEMLIMIT"] = "512MiB"
	values["HARPY_RABBITMQ_URL_FILE"] = secretRoot + "rabbitmq-url.txt"
	addSharedStorageEnvironment(set, values)
	addSharedKeyringEnvironment(set, values)
	addS3Environment(set, values, "HARPY_DOCUMENTS_LEGACY_SOURCE",
		"harpy-documents-legacy")
	publicURL := "http://localhost:5173"
	if set.options.profile == profileForgeGuard {
		publicURL = set.options.publicURL
	}
	for key, value := range (profileEnvironment{
		"HARPY_DOCUMENTS_ARCHIVE_RESERVATION_TTL": "24h",
		"HARPY_DOCUMENTS_CLAMAV_ADDRESS":           "clamav:3310",
		"HARPY_AI_RUNTIME_CONFIG_FILE":             secretRoot + "ai-runtime.json",
		"HARPY_MCP_RUNTIME_CONFIG_FILE":            secretRoot + "mcp-runtime.json",
		"HARPY_NOTIFICATIONS_KEK_KEYRING_FILE":    secretRoot + "notifications-kek-keyring.json",
		"HARPY_FORMS_EXPORT_SPOOL_DIR":            "/var/lib/harpy/forms-export-spool",
		"HARPY_ANALYTICS_DASHBOARD_EXPORT_SPOOL_DIR": "/var/lib/harpy/analytics-dashboard-export-spool",
		"HARPY_IAM_INVITATION_SMTP_HOST":          "mailpit",
		"HARPY_IAM_INVITATION_SMTP_PORT":          "1025",
		"HARPY_IAM_INVITATION_SMTP_USERNAME":      "harpy-local",
		"HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE": secretRoot + "smtp-password.txt",
		"HARPY_IAM_INVITATION_FROM_ADDRESS":       "security@local.harpy.test",
		"HARPY_IAM_INVITATION_FROM_NAME":          "ForgeCRUD Security",
		"HARPY_IAM_INVITATION_PUBLIC_BASE_URL":    publicURL,
		"HARPY_IAM_INVITATION_TEMPLATE_FILE":      secretRoot + "invitation-template.json",
		"HARPY_IAM_INVITATION_SMTP_TIMEOUT":       "5s",
	}) {
		values[key] = value
	}
	if set.options.profile == profileForgeGuard {
		values["HARPY_MANAGEMENT_LISTEN_ADDR"] = "0.0.0.0:9000"
		values["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] =
			"/run/recovery/recovery-trust-bundle.v1.json"
		values["HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"] =
			secretRoot + "protected-authority-ack-trust-bundle.v1.json"
		values["HARPY_LICENSE_TRUST_BUNDLE_PATH"] =
			secretRoot + "license-trust-bundle.v1.json"
		addForgeGuardLicenseEnvironment(values, set.options)
		values["HARPY_RELEASE_TRUST_ROOT_PATH"] =
			secretRoot + "protected-authority-release-trust-root.json"
	} else {
		for key, value := range localWorkerAuthorityEnvironment() {
			values[key] = value
		}
	}
	return values
}
