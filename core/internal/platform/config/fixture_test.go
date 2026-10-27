package config

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	validIAMTOTPKMSKeyID = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7"
	validIAMDeliveryKMSKeyID = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/4"
)

func validEnvironment() []string {
	return []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN=postgresql://user@example.invalid/harpy?sslmode=disable",
	}
}

func validAPIEnvironment() []string {
	return validAPIEnvironmentAt(time.Now().UTC().Truncate(time.Second))
}

func validWorkerEnvironment() []string {
	return validWorkerEnvironmentAt(time.Now().UTC().Truncate(time.Second))
}

func validWorkerEnvironmentAt(now time.Time) []string {
	entries := append(validEnvironment(),
		"GOMEMLIMIT=512MiB",
		"HARPY_RABBITMQ_URL=amqp://worker:secret@rabbit.invalid/%2f",
		cursorKeyringEnvironmentEntry(),
		catalogPackageKeyringEnvironmentEntry(),
	)
	entries = append(entries, validFormsExportEnvironment(true)...)
	entries = append(entries, validFormsImportSourceEnvironment(false)...)
	entries = append(entries, validAnalyticsDashboardExportEnvironment(true)...)
	entries = append(entries, validIAMSystemTransferEnvironment()...)
	entries = append(entries, validDocumentsEnvironment(true)...)
	entries = append(entries, validNotificationsKeyringEnvironmentEntry())
	entries = append(entries, validWorkerDevelopmentAuthorityEnvironment()...)
	directory, err := os.MkdirTemp("", "harpy-worker-config-")
	if err != nil {
		panic(err)
	}
	passwordPath := filepath.Join(directory, "smtp-password.txt")
	templatePath := filepath.Join(directory, "invitation-template.json")
	if err := os.WriteFile(passwordPath, []byte("fixture-password"), 0o600); err != nil {
		panic(err)
	}
	template := `{"version":1,"subject":"Invitation","text":"Open {{INVITATION_URL}}",` +
		`"html":"<a href=\"{{INVITATION_URL}}\">Open</a>"}`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		panic(err)
	}
	return append(entries,
		validIAMDirectoryEnvironmentEntry(now),
		iamOneTimeKeyringEnvironmentKey+"="+
			validIAMHMACKeyringJSON(now, iamPurposeOneTime, 25),
		iamDeliveryKMSKeyIDEnvironmentKey+"="+validIAMDeliveryKMSKeyID,
		"HARPY_IAM_INVITATION_SMTP_HOST=smtp.example.invalid",
		"HARPY_IAM_INVITATION_SMTP_PORT=587",
		"HARPY_IAM_INVITATION_SMTP_USERNAME=worker",
		"HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE="+passwordPath,
		"HARPY_IAM_INVITATION_FROM_ADDRESS=security@example.invalid",
		"HARPY_IAM_INVITATION_FROM_NAME=Harpy Security",
		"HARPY_IAM_INVITATION_PUBLIC_BASE_URL=https://app.example.invalid",
		"HARPY_IAM_INVITATION_TEMPLATE_FILE="+templatePath,
		"HARPY_IAM_INVITATION_SMTP_TIMEOUT=5s",
	)
}

func validAPIEnvironmentAt(now time.Time) []string {
	entries := append(validEnvironment(), cursorKeyringEnvironmentEntry(),
		catalogPackageKeyringEnvironmentEntry(),
		deploymentModeEnvironmentKey+"=saas")
	entries = append(entries, validFormsExportEnvironment(false)...)
	entries = append(entries, validFormsImportSourceEnvironment(true)...)
	entries = append(entries, validAnalyticsDashboardExportEnvironment(false)...)
	entries = append(entries, validIAMSystemTransferEnvironment()...)
	entries = append(entries, validDocumentsEnvironment(false)...)
	entries = append(entries, validNotificationsConsumerKeyringEnvironmentEntry())
	entries = append(entries, validAPISaaSAuthorityEnvironment()...)
	return append(entries, validIAMAPIEnvironment(now)...)
}

func catalogPackageKeyringEnvironmentEntry() string {
	key := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{7}, ed25519.PublicKeySize))
	return catalogPackageKeyringEnvironmentKey + `={"version":1,"keys":[{"key_id":"release-2026-07","public_key":"` + key + `"}]}`
}

func cursorKeyringEnvironmentEntry() string {
	return cursorKeyringEnvironmentKey + "=" + validCursorKeyringJSON()
}

func validCursorKeyringJSON() string {
	key := func(value byte) string {
		return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, cursorKeySize))
	}
	return fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":"%s"},"previous":{"kid":"previous-1","key":"%s"}}`,
		key(1), key(2),
	)
}

func validIAMAPIEnvironment(now time.Time) []string {
	entries := []string{
		iamAccessIssuerEnvironmentKey + "=https://api.harpy.invalid",
		iamAccessAudienceEnvironmentKey + "=https://api.harpy.invalid/api",
		iamAllowedOriginsEnvironmentKey + `=["https://app.harpy.invalid"]`,
		iamAccessKeyringEnvironmentKey + "=" + validIAMAccessKeyringJSON(now, 10),
		iamPasswordBlocklistsEnvironmentKey + "=" + validIAMPasswordBlocklistsJSON(),
	}
	purposes := iamAPIPurposes()
	keys := []string{
		iamRefreshKeyringEnvironmentKey, iamCSRFKeyringEnvironmentKey,
		iamDirectoryKeyringEnvironmentKey, iamThrottleKeyringEnvironmentKey,
		iamSessionKeyringEnvironmentKey, iamOneTimeKeyringEnvironmentKey,
		iamRecoveryKeyringEnvironmentKey,
	}
	for index, purpose := range purposes {
		entries = append(entries, keys[index]+"="+validIAMHMACKeyringJSON(now, purpose, byte(20+index)))
	}
	entries = append(entries,
		iamTOTPKMSKeyIDEnvironmentKey+"="+validIAMTOTPKMSKeyID,
		iamDeliveryKMSKeyIDEnvironmentKey+"="+validIAMDeliveryKMSKeyID,
	)
	return append(entries, validIAMActorBridgeEnvironment()...)
}

func validIAMDirectoryEnvironmentEntry(now time.Time) string {
	return iamDirectoryKeyringEnvironmentKey + "=" +
		validIAMHMACKeyringJSON(now, iamPurposeDirectory, 22)
}

func validIAMAccessKeyringJSON(now time.Time, marker byte) string {
	activatedAt := now.Add(-24 * time.Hour)
	return fmt.Sprintf(
		`{"version":1,"active":{"kid":%q,"seed":%q,"activated_at":%q,"sign_until":%q},"retired_public":[]}`,
		iamFixtureKID(marker), iamFixtureMaterial(marker), iamFixtureTime(activatedAt),
		iamFixtureTime(activatedAt.Add(30*24*time.Hour)),
	)
}

func validIAMHMACKeyringJSON(now time.Time, purpose iamHMACPurpose, marker byte) string {
	activatedAt := now.Add(-24 * time.Hour)
	return fmt.Sprintf(
		`{"version":1,"purpose":%q,"active":{"kid":%q,"key":%q,"activated_at":%q,"issue_until":%q},"retained":[]}`,
		purpose, iamFixtureKID(marker), iamFixtureMaterial(marker),
		iamFixtureTime(activatedAt), iamFixtureTime(activatedAt.Add(90*24*time.Hour)),
	)
}

func iamFixtureKID(marker byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 12))
}

func iamFixtureMaterial(marker byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 32))
}

func iamFixtureTime(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05Z")
}

func withEnvironment(entries []string, replacements ...string) []string {
	values := append([]string(nil), entries...)
	for _, replacement := range replacements {
		key := replacement
		for index, character := range replacement {
			if character == '=' {
				key = replacement[:index]
				break
			}
		}
		replaced := false
		for index, existing := range values {
			if len(existing) > len(key) && existing[:len(key)+1] == key+"=" {
				values[index] = replacement
				replaced = true
				break
			}
		}
		if !replaced {
			values = append(values, replacement)
		}
	}
	return values
}
