package main

func workerEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(secretRoot + "postgres-runtime-dsn.txt")
	values["GOMEMLIMIT"] = "512MiB"
	values["HARPY_RABBITMQ_URL_FILE"] = secretRoot + "rabbitmq-url.txt"
	addSharedStorageEnvironment(values)
	addSharedKeyringEnvironment(values)
	addS3Environment(values, "HARPY_DOCUMENTS_LEGACY_SOURCE",
		"harpy-documents-legacy")
	for key, value := range (profileEnvironment{
		"HARPY_DOCUMENTS_ARCHIVE_RESERVATION_TTL":    "24h",
		"HARPY_DOCUMENTS_CLAMAV_ADDRESS":             "clamav:3310",
		"HARPY_NOTIFICATIONS_KEK_KEYRING_FILE":       secretRoot + "notifications-kek-keyring.json",
		"HARPY_ANALYTICS_DASHBOARD_EXPORT_SPOOL_DIR": "/var/lib/harpy/analytics-dashboard-export-spool",
		"HARPY_IAM_INVITATION_SMTP_HOST":             "mailpit",
		"HARPY_IAM_INVITATION_SMTP_PORT":             "1025",
		"HARPY_IAM_INVITATION_SMTP_USERNAME":         "harpy-local",
		"HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE":    secretRoot + "smtp-password.txt",
		"HARPY_IAM_INVITATION_FROM_ADDRESS":          "security@local.harpy.test",
		"HARPY_IAM_INVITATION_FROM_NAME":             "Harpy Security",
		"HARPY_IAM_INVITATION_PUBLIC_BASE_URL":       "http://localhost:5173",
		"HARPY_IAM_INVITATION_TEMPLATE_FILE":         secretRoot + "invitation-template.json",
		"HARPY_IAM_INVITATION_SMTP_TIMEOUT":          "5s",
	}) {
		values[key] = value
	}
	for key, value := range localWorkerAuthorityEnvironment() {
		values[key] = value
	}
	return values
}
