package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIAMAPIProfileLoadsExactPurposeSet(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	loaded, err := loadAPIAt(validAPIEnvironmentAt(now), now)
	if err != nil {
		t.Fatal(err)
	}
	iamConfig, err := loaded.IAM()
	if err != nil || iamConfig.Validate() != nil {
		t.Fatalf("IAM API config unavailable: %v", err)
	}
	access, _ := iamConfig.AccessTokenKeyring()
	refresh, _ := iamConfig.RefreshTokenKeyring()
	directory, _ := iamConfig.LoginDirectoryKeyring()
	if access.Validate() != nil || refresh.Validate() != nil || directory.Validate() != nil {
		t.Fatal("loaded IAM purpose keyring is invalid")
	}
	totp, _ := iamConfig.TOTPKMSKeyID()
	delivery, _ := iamConfig.DeliveryKMSKeyID()
	if value, _ := totp.Value(); value != validIAMTOTPKMSKeyID {
		t.Fatalf("unexpected TOTP key ID %q", value)
	}
	if value, _ := delivery.Value(); value != validIAMDeliveryKMSKeyID {
		t.Fatalf("unexpected delivery key ID %q", value)
	}
}

func TestIAMProfilesFailClosedOnMissingKeys(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	for _, key := range []string{
		iamAccessIssuerEnvironmentKey, iamAccessAudienceEnvironmentKey,
		iamAllowedOriginsEnvironmentKey,
		iamAccessKeyringEnvironmentKey, iamRefreshKeyringEnvironmentKey,
		iamCSRFKeyringEnvironmentKey, iamDirectoryKeyringEnvironmentKey,
		iamThrottleKeyringEnvironmentKey, iamSessionKeyringEnvironmentKey,
		iamOneTimeKeyringEnvironmentKey, iamRecoveryKeyringEnvironmentKey,
		iamPasswordBlocklistsEnvironmentKey,
	} {
		t.Run(key, func(t *testing.T) {
			entries := withoutEnvironmentKey(validAPIEnvironmentAt(now), key)
			if _, err := loadAPIAt(entries, now); err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("missing key was not rejected safely: %v", err)
			}
		})
	}
	worker := withoutEnvironmentKey(validWorkerEnvironmentAt(now), iamDirectoryKeyringEnvironmentKey)
	if _, err := loadWorkerAt(worker, now); err == nil {
		t.Fatal("worker started without directory keyring")
	}
}

func TestIAMProfilesEnforceProcessSecretBoundary(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	worker := append(validWorkerEnvironmentAt(now),
		iamAccessKeyringEnvironmentKey+"="+validIAMAccessKeyringJSON(now, 10))
	if _, err := loadWorkerAt(worker, now); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("worker accepted access signing material: %v", err)
	}
	if _, err := loadSchemaExecutor(append(validEnvironment(),
		validIAMDirectoryEnvironmentEntry(now))); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("schema executor accepted IAM material: %v", err)
	}
	rawKey := "HARPY_IAM_TOTP_KEY=" + iamFixtureMaterial(31)
	if _, err := loadAPIAt(append(validAPIEnvironmentAt(now), rawKey), now); err == nil {
		t.Fatal("API accepted raw TOTP key environment surface")
	}
}

func TestIAMAPIRejectsCrossPurposeKIDAndMaterialReuse(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	csrf := validIAMHMACKeyringJSON(now, iamPurposeCSRF, 21)
	tests := map[string]string{
		"kid and material": validIAMHMACKeyringJSON(now, iamPurposeCSRF, 20),
		"kid":              strings.Replace(csrf, iamFixtureKID(21), iamFixtureKID(20), 1),
		"material": strings.Replace(
			csrf, iamFixtureMaterial(21), iamFixtureMaterial(20), 1,
		),
	}
	for name, document := range tests {
		t.Run(name, func(t *testing.T) {
			entries := withEnvironment(
				validAPIEnvironmentAt(now), iamCSRFKeyringEnvironmentKey+"="+document,
			)
			if _, err := loadAPIAt(entries, now); err == nil ||
				!strings.Contains(err.Error(), "iam-key-purpose-isolation.v1") {
				t.Fatalf("cross-purpose reuse accepted: %v", err)
			}
		})
	}
}

func TestIAMKeyringFileXORAndRedactedErrors(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	secretValue := validIAMHMACKeyringJSON(now, iamPurposeRefresh, 20)
	path := filepath.Join(t.TempDir(), "refresh.json")
	if err := os.WriteFile(path, []byte(secretValue+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := withoutEnvironmentKey(validAPIEnvironmentAt(now), iamRefreshKeyringEnvironmentKey)
	entries = append(entries, iamRefreshKeyringEnvironmentKey+"_FILE="+path)
	if _, err := loadAPIAt(entries, now); err != nil {
		t.Fatalf("file-only IAM keyring rejected: %v", err)
	}
	both := append(validAPIEnvironmentAt(now), iamRefreshKeyringEnvironmentKey+"_FILE="+path)
	if _, err := loadAPIAt(both, now); err == nil || strings.Contains(err.Error(), secretValue) {
		t.Fatalf("IAM XOR error missing or leaked material: %v", err)
	}
}

func TestIAMKMSKeyIDRejectsAliasesAndMalformedVersions(t *testing.T) {
	for _, value := range []string{
		"", "kms://iam/totp/current", validIAMTOTPKMSKeyID + "/",
		validIAMTOTPKMSKeyID + "?x=1", strings.Repeat("x", 513),
	} {
		if _, ok := parseIAMKMSKeyID(value); ok {
			t.Fatalf("invalid KMS key ID accepted: %q", value)
		}
	}
}

func TestIAMAPIRequiresDistinctKMSPurposeIDs(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	sharedKeyVersion := strings.Replace(validIAMTOTPKMSKeyID,
		"cryptoKeyVersions/7", "cryptoKeyVersions/6", 1)
	for _, keyID := range []string{validIAMTOTPKMSKeyID, sharedKeyVersion} {
		entries := withEnvironment(
			validAPIEnvironmentAt(now),
			iamDeliveryKMSKeyIDEnvironmentKey+"="+keyID,
		)
		if _, err := loadAPIAt(entries, now); err == nil ||
			!strings.Contains(err.Error(), "iam-key-purpose-isolation.v1") {
			t.Fatalf("shared TOTP/delivery KMS CryptoKey accepted: %v", err)
		}
	}
}

func TestIAMInvalidDocumentErrorNamesOnlyKeyAndContract(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	secretMarker := "must-not-leak-key-material"
	entries := withEnvironment(
		validAPIEnvironmentAt(now),
		iamAccessKeyringEnvironmentKey+"="+secretMarker,
	)
	_, err := loadAPIAt(entries, now)
	if err == nil || !strings.Contains(err.Error(), iamAccessKeyringEnvironmentKey) ||
		!strings.Contains(err.Error(), "iam-access-token-keyring.v1") ||
		strings.Contains(err.Error(), secretMarker) {
		t.Fatalf("invalid IAM document error is not stable and redacted: %v", err)
	}
}
