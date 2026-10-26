package runtimev1

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
)

func addIAMAPIProfileSnapshot(variables map[string]any) {
	variables["HARPY_IAM_ACCESS_TOKEN_KEYRING"] = iamAccessContractValue(10)
	purposes := []struct {
		key     string
		purpose string
		marker  byte
	}{
		{"HARPY_IAM_REFRESH_TOKEN_KEYRING", "refresh_token_v1", 20},
		{"HARPY_IAM_CSRF_KEYRING", "csrf_v1", 21},
		{"HARPY_IAM_LOGIN_DIRECTORY_KEYRING", "login_directory_v1", 22},
		{"HARPY_IAM_AUTH_THROTTLE_KEYRING", "auth_throttle_v1", 23},
		{"HARPY_IAM_SESSION_NETWORK_KEYRING", "session_network_v1", 24},
		{"HARPY_IAM_ONE_TIME_TOKEN_KEYRING", "one_time_token_v1", 25},
		{"HARPY_IAM_RECOVERY_CODE_KEYRING", "recovery_code_v1", 26},
	}
	for _, purpose := range purposes {
		variables[purpose.key] = iamHMACContractValue(purpose.purpose, purpose.marker)
	}
	variables["HARPY_IAM_TOTP_KMS_KEY_ID"] =
		"gcp-kms://projects/harpy-prod/locations/europe-west1/" +
			"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7"
	variables["HARPY_IAM_DELIVERY_KMS_KEY_ID"] =
		"gcp-kms://projects/harpy-prod/locations/europe-west1/" +
			"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/4"
	variables["HARPY_IAM_PASSWORD_BLOCKLISTS"] = fmt.Sprintf(
		`{"version":1,"normalization":"unicode_nfc_v1","digest":"sha256","common":{"path":"/artifacts/common-v1.bin","sha256":%q},"compromised":{"path":"/artifacts/compromised-v1.bin","sha256":%q}}`,
		strings.Repeat("1", 64), strings.Repeat("2", 64),
	)
	variables["HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_KEY_ID"] = "edge-2026-07"
	variables["HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_SEED_FILE"] =
		"/secrets/iam-v1-actor-signing-seed"
	variables["HARPY_IAM_V1_ACTOR_BRIDGE_PUBLIC_KEYS_FILE"] =
		"/config/iam-v1-actor-public-keys.json"
	variables["HARPY_IAM_V1_ACTOR_BRIDGE_TARGETS_FILE"] =
		"/config/iam-v1-actor-targets.json"
	variables["HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_CERT_FILE"] =
		"/secrets/iam-v1-actor-client.crt"
	variables["HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_KEY_FILE"] =
		"/secrets/iam-v1-actor-client.key"
	variables["HARPY_IAM_V1_ACTOR_BRIDGE_SERVER_CA_FILE"] =
		"/config/iam-v1-actor-server-ca.crt"
}

func addIAMWorkerProfileSnapshot(variables map[string]any) {
	variables["HARPY_IAM_LOGIN_DIRECTORY_KEYRING"] =
		iamHMACContractValue("login_directory_v1", 22)
	variables["HARPY_IAM_ONE_TIME_TOKEN_KEYRING"] =
		iamHMACContractValue("one_time_token_v1", 25)
	variables["HARPY_IAM_DELIVERY_KMS_KEY_ID"] =
		"gcp-kms://projects/harpy-prod/locations/europe-west1/" +
			"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/4"
	variables["HARPY_IAM_INVITATION_SMTP_HOST"] = "smtp.example.invalid"
	variables["HARPY_IAM_INVITATION_SMTP_PORT"] = "587"
	variables["HARPY_IAM_INVITATION_SMTP_USERNAME"] = "worker"
	variables["HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE"] = "/secrets/smtp-password"
	variables["HARPY_IAM_INVITATION_FROM_ADDRESS"] = "security@example.invalid"
	variables["HARPY_IAM_INVITATION_FROM_NAME"] = "Harpy Security"
	variables["HARPY_IAM_INVITATION_PUBLIC_BASE_URL"] = "https://app.example.invalid"
	variables["HARPY_IAM_INVITATION_TEMPLATE_FILE"] = "/config/invitation-template.json"
	variables["HARPY_IAM_INVITATION_SMTP_TIMEOUT"] = "10s"
}

func iamAccessContractValue(marker byte) string {
	return fmt.Sprintf(
		`{"version":1,"active":{"kid":%q,"seed":%q,"activated_at":"2026-07-01T00:00:00Z","sign_until":"2026-07-31T00:00:00Z"},"retired_public":[]}`,
		iamContractKID(marker), iamContractMaterial(marker),
	)
}

func iamHMACContractValue(purpose string, marker byte) string {
	return fmt.Sprintf(
		`{"version":1,"purpose":%q,"active":{"kid":%q,"key":%q,"activated_at":"2026-07-01T00:00:00Z","issue_until":"2026-09-29T00:00:00Z"},"retained":[]}`,
		purpose, iamContractKID(marker), iamContractMaterial(marker),
	)
}

func iamContractKID(marker byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 12))
}

func iamContractMaterial(marker byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 32))
}
