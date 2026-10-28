package main

const (
	localTOTPKMSKeyID = "gcp-kms://projects/harpy-local/locations/global/" +
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/1"
	localDeliveryKMSKeyID = "gcp-kms://projects/harpy-local/locations/global/" +
		"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/1"
	hostTOTPKMSKeyID = "gcp-kms://projects/harpy-host/locations/global/" +
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/1"
	hostDeliveryKMSKeyID = "gcp-kms://projects/harpy-host/locations/global/" +
		"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/1"
)

func apiEnvironment(set *artifactSet) profileEnvironment {
	values := baseRuntimeEnvironmentFor(set, secretRoot+"postgres-runtime-dsn.txt")
	addSharedStorageEnvironment(set, values)
	addSharedKeyringEnvironment(set, values)
	issuer := "http://localhost:8000"
	origins := `["http://localhost:5173","http://127.0.0.1:5173"]`
	totpKey := localTOTPKMSKeyID
	if set.options.profile == profileForgeGuard {
		issuer = set.options.publicURL
		origins = `["` + set.options.publicURL + `"]`
		totpKey = hostTOTPKMSKeyID
	}
	for key, value := range (profileEnvironment{
		"HARPY_DOCUMENTS_CLAMAV_ADDRESS":              "clamav:3310",
		"HARPY_AI_RUNTIME_CONFIG_FILE":                secretRoot + "ai-runtime.json",
		"HARPY_MCP_RUNTIME_CONFIG_FILE":               secretRoot + "mcp-runtime.json",
		"HARPY_AI_SYSTEM_KNOWLEDGE_DIR":               "/var/lib/harpy/ai-system-knowledge",
		"HARPY_NOTIFICATIONS_CONSUMER_KEYRING_FILE":   secretRoot + "notifications-consumer-keyring.json",
		"HARPY_FORMS_IMPORT_SOURCE_STAGING_DIR":       "/var/lib/harpy/forms-import-staging",
		"HARPY_IAM_ACCESS_TOKEN_ISSUER":               issuer,
		"HARPY_IAM_ACCESS_TOKEN_AUDIENCE":             issuer + "/api",
		"HARPY_IAM_ALLOWED_ORIGINS":                   origins,
		"HARPY_IAM_ACCESS_TOKEN_KEYRING_FILE":         secretRoot + "iam-access-token-keyring.json",
		"HARPY_IAM_REFRESH_TOKEN_KEYRING_FILE":        secretRoot + "iam-refresh-token-keyring.json",
		"HARPY_IAM_CSRF_KEYRING_FILE":                 secretRoot + "iam-csrf-keyring.json",
		"HARPY_IAM_AUTH_THROTTLE_KEYRING_FILE":        secretRoot + "iam-auth-throttle-keyring.json",
		"HARPY_IAM_SESSION_NETWORK_KEYRING_FILE":      secretRoot + "iam-session-network-keyring.json",
		"HARPY_IAM_RECOVERY_CODE_KEYRING_FILE":        secretRoot + "iam-recovery-code-keyring.json",
		"HARPY_IAM_PASSWORD_BLOCKLISTS_FILE":          secretRoot + "iam-password-blocklists.json",
		"HARPY_IAM_TOTP_KMS_KEY_ID":                   totpKey,
		"HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_KEY_ID":    set.actorBridgeKeyID,
		"HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_SEED_FILE": secretRoot + "actor-bridge-signing-seed",
		"HARPY_IAM_V1_ACTOR_BRIDGE_PUBLIC_KEYS_FILE":  secretRoot + "actor-bridge-public-keys.json",
		"HARPY_IAM_V1_ACTOR_BRIDGE_TARGETS_FILE":      secretRoot + "actor-bridge-targets.json",
		"HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_CERT_FILE":  secretRoot + "actor-bridge-client.crt",
		"HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_KEY_FILE":   secretRoot + "actor-bridge-client.key",
		"HARPY_IAM_V1_ACTOR_BRIDGE_SERVER_CA_FILE":    secretRoot + "actor-bridge-server-ca.crt",
	}) {
		values[key] = value
	}
	if set.options.profile == profileForgeGuard {
		values["HARPY_PUBLIC_LISTEN_ADDR"] = "0.0.0.0:8000"
		values["HARPY_MANAGEMENT_LISTEN_ADDR"] = "0.0.0.0:9000"
		values["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] =
			"/run/recovery/recovery-trust-bundle.v1.json"
		values["HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE"] =
			"/run/recovery/recovery-control-signer.ed25519"
		values["HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"] =
			secretRoot + "protected-authority-ack-trust-bundle.v1.json"
	} else {
		for key, value := range localAPIAuthorityEnvironment() {
			values[key] = value
		}
	}
	return values
}
