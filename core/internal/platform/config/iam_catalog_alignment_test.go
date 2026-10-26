package config

import "testing"

func TestPublishedIAMPurposesMatchRuntimeBindings(t *testing.T) {
	expected := map[string]string{
		iamAccessKeyringEnvironmentKey:    "access_token_v1",
		iamRefreshKeyringEnvironmentKey:   string(iamPurposeRefresh),
		iamCSRFKeyringEnvironmentKey:      string(iamPurposeCSRF),
		iamDirectoryKeyringEnvironmentKey: string(iamPurposeDirectory),
		iamThrottleKeyringEnvironmentKey:  string(iamPurposeThrottle),
		iamSessionKeyringEnvironmentKey:   string(iamPurposeSessionNetwork),
		iamOneTimeKeyringEnvironmentKey:   string(iamPurposeOneTime),
		iamRecoveryKeyringEnvironmentKey:  string(iamPurposeRecovery),
		iamTOTPKMSKeyIDEnvironmentKey:     "totp_kms_v1",
		iamDeliveryKMSKeyIDEnvironmentKey: "delivery_kms_v1",
	}
	seen := make(map[string]bool, len(expected))
	for _, variable := range readCatalog(t).Variables {
		purpose, exists := expected[variable.Name]
		if !exists {
			continue
		}
		if variable.KeyPurpose != purpose {
			t.Fatalf("%s purpose=%q want=%q", variable.Name, variable.KeyPurpose, purpose)
		}
		seen[variable.Name] = true
	}
	for key := range expected {
		if !seen[key] {
			t.Fatalf("published IAM config key is missing: %s", key)
		}
	}
}
