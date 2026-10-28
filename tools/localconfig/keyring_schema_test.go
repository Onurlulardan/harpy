package main

import "testing"

func TestGeneratedTypedKeyringsSatisfyRuntimeSchemas(t *testing.T) {
	set, _ := generatedSetForTest(t)
	tests := []struct{ file, schema string }{
		{"rest-query-cursor-keyring.json", "rest-query-cursor-keyring.schema.json"},
		{"catalog-package-public-keyring.json", "catalog-package-public-keyring.schema.json"},
		{"documents-kek-keyring.json", "documents-kek-keyring.schema.json"},
		{"analytics-dashboard-export-fcart-kek-keyring.json", "analytics-dashboard-export-fcart-kek-keyring.schema.json"},
		{"forms-export-fcart-kek-keyring.json", "forms-export-fcart-kek-keyring.schema.json"},
		{"forms-import-source-fcart-kek-keyring.json", "forms-import-source-fcart-kek-keyring.schema.json"},
		{"iam-access-token-keyring.json", "iam-access-token-keyring.schema.json"},
		{"iam-refresh-token-keyring.json", "iam-hmac-keyring.schema.json"},
		{"iam-csrf-keyring.json", "iam-hmac-keyring.schema.json"},
		{"iam-login-directory-keyring.json", "iam-hmac-keyring.schema.json"},
		{"iam-auth-throttle-keyring.json", "iam-hmac-keyring.schema.json"},
		{"iam-session-network-keyring.json", "iam-hmac-keyring.schema.json"},
		{"iam-one-time-token-keyring.json", "iam-hmac-keyring.schema.json"},
		{"iam-recovery-code-keyring.json", "iam-hmac-keyring.schema.json"},
		{"iam-password-blocklists.json", "iam-password-blocklists.schema.json"},
		{"iam-system-export-fcart-kek-keyring.json", "iam-system-transfer-fcart-kek-keyring.schema.json"},
		{"iam-system-import-source-fcart-kek-keyring.json", "iam-system-transfer-fcart-kek-keyring.schema.json"},
		{"notifications-consumer-keyring.json", "notifications-consumer-keyring.schema.json"},
		{"notifications-kek-keyring.json", "notifications-kek-keyring.schema.json"},
	}
	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			value := decodeJSONArtifact(t, set.files[test.file])
			if err := compileRuntimeSchema(t, test.schema).Validate(value); err != nil {
				t.Fatalf("generated keyring rejected: %v", err)
			}
		})
	}
}
