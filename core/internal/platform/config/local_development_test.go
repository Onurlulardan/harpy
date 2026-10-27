package config

import (
	"strings"
	"testing"
	"time"
)

func TestDevelopmentOmitsClamAVOwnerAndKMS(t *testing.T) {
	now := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	entries := validAPIEnvironmentAt(now)
	entries = withoutEnvironmentKey(entries, documentsOwnerKey)
	entries = withoutEnvironmentKey(entries, documentsClamAVAddressKey)
	entries = withoutEnvironmentKey(entries, iamTOTPKMSKeyIDEnvironmentKey)
	entries = withoutEnvironmentKey(entries, iamDeliveryKMSKeyIDEnvironmentKey)
	entries = withoutEnvironmentKey(entries, saasKMSKeyIDKey)
	loaded, err := loadAPIAt(entries, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := loaded.IAM(); err != nil {
		t.Fatal(err)
	}
}

func TestProductionStillRequiresIAMKMSAndBucketOwner(t *testing.T) {
	_, err := loadIAMKMSKeyID(environmentValues{}, iamTOTPKMSKeyIDEnvironmentKey, EnvironmentProduction)
	if err == nil || !strings.Contains(err.Error(), iamTOTPKMSKeyIDEnvironmentKey) {
		t.Fatalf("production KMS error = %v", err)
	}
	_, err = loadIAMKMSKeyID(environmentValues{}, iamTOTPKMSKeyIDEnvironmentKey, EnvironmentDevelopment)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = resolveBucketOwner(environmentValues{}, documentsOwnerKey, EnvironmentProduction, true)
	if err == nil || !strings.Contains(err.Error(), documentsOwnerKey) {
		t.Fatalf("production owner error = %v", err)
	}
	_, optional, err := resolveBucketOwner(environmentValues{}, documentsOwnerKey, EnvironmentDevelopment, true)
	if err != nil || !optional {
		t.Fatalf("development owner = %t, %v", optional, err)
	}
}

func TestUnknownHarpyKeyIncludesItsName(t *testing.T) {
	_, err := loadAPI(append(validEnvironment(), "HARPY_TYPO=value"))
	if err == nil || !strings.Contains(err.Error(), "HARPY_TYPO") || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("error = %v", err)
	}
}
