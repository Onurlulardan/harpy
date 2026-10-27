package config

func runtimeKeys() []string {
	return []string{
		"HARPY_ENVIRONMENT", "HARPY_POSTGRES_ENDPOINT_KIND",
		"HARPY_POSTGRES_DSN", "HARPY_POSTGRES_DSN_FILE",
		"HARPY_POSTGRES_MAX_CONNS", "HARPY_POSTGRES_MIN_CONNS",
		"HARPY_POSTGRES_CONNECT_TIMEOUT", "HARPY_POSTGRES_CLEANUP_TIMEOUT",
		"HARPY_POSTGRES_MAX_CONN_LIFETIME", "HARPY_POSTGRES_MAX_CONN_IDLE_TIME",
		"HARPY_POSTGRES_HEALTH_CHECK_PERIOD", "HARPY_READINESS_TIMEOUT",
		"HARPY_LOG_LEVEL", "HARPY_OTLP_ENDPOINT", "HARPY_OTLP_HEADERS",
		"HARPY_OTLP_HEADERS_FILE", "HARPY_OTLP_INSECURE",
		"HARPY_OTLP_EXPORT_TIMEOUT", "HARPY_TRACE_SAMPLE_RATIO",
	}
}

func longRunningKeys() []string {
	return []string{"HARPY_MANAGEMENT_LISTEN_ADDR", "HARPY_DRAIN_TIMEOUT"}
}

func apiKeys() []string {
	keys := []string{
		deploymentModeEnvironmentKey,
		"HARPY_PUBLIC_LISTEN_ADDR", "HARPY_HTTP_READ_HEADER_TIMEOUT",
		"HARPY_HTTP_IDLE_TIMEOUT", "HARPY_HTTP_UNARY_TIMEOUT",
		"HARPY_HTTP_MAX_HEADER_BYTES", "HARPY_HTTP_MAX_JSON_BODY_BYTES",
		analyticsTrustedProxyKey,
		cursorKeyringEnvironmentKey, cursorKeyringEnvironmentKey + "_FILE",
		catalogPackageKeyringEnvironmentKey, catalogPackageKeyringEnvironmentKey + "_FILE",
		notificationsConsumerKeyringFileKey,
	}
	keys = append(keys, analyticsDashboardExportArtifactKeys()...)
	keys = append(keys, iamSystemTransferArtifactKeys()...)
	keys = append(keys, documentsArtifactKeys()...)
	keys = append(keys, documentsClamAVAddressKey)
	return append(keys, iamAPIKeys()...)
}

func workerKeys() []string {
	keys := []string{
		workerMemoryLimitKey,
		deploymentModeEnvironmentKey,
		"HARPY_RABBITMQ_URL", "HARPY_RABBITMQ_URL_FILE",
		cursorKeyringEnvironmentKey, cursorKeyringEnvironmentKey + "_FILE",
		catalogPackageKeyringEnvironmentKey, catalogPackageKeyringEnvironmentKey + "_FILE",
		notificationsKEKKeyringFileKey,
		"HARPY_IAM_INVITATION_SMTP_HOST",
		"HARPY_IAM_INVITATION_SMTP_PORT",
		"HARPY_IAM_INVITATION_SMTP_USERNAME",
		"HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE",
		"HARPY_IAM_INVITATION_FROM_ADDRESS",
		"HARPY_IAM_INVITATION_FROM_NAME",
		"HARPY_IAM_INVITATION_PUBLIC_BASE_URL",
		"HARPY_IAM_INVITATION_TEMPLATE_FILE",
		"HARPY_IAM_INVITATION_SMTP_TIMEOUT",
	}
	keys = append(keys, analyticsDashboardExportArtifactKeys()...)
	keys = append(keys, analyticsExportSpoolKey)
	keys = append(keys, iamSystemTransferArtifactKeys()...)
	keys = append(keys, documentsArtifactKeys()...)
	keys = append(keys, documentsLegacyArtifactKeys()...)
	keys = append(keys, documentsClamAVAddressKey)
	return append(keys, iamWorkerKeys()...)
}
