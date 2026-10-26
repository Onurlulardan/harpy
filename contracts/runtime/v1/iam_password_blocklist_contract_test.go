package runtimev1

import (
	"strings"
	"testing"
)

func TestIAMPasswordBlocklistSchemaPinsOfflineCoordinates(t *testing.T) {
	schema := compileSchema(t, "iam-password-blocklists.schema.json")
	valid := iamPasswordBlocklistContractDocument()
	if err := schema.Validate(valid); err != nil {
		t.Fatalf("valid blocklist coordinates rejected: %v", err)
	}
	tests := map[string]func(map[string]any){
		"version": func(value map[string]any) { value["version"] = 2 },
		"profile": func(value map[string]any) { value["normalization"] = "NFKC" },
		"digest": func(value map[string]any) {
			value["common"].(map[string]any)["sha256"] = strings.Repeat("A", 64)
		},
		"relative path": func(value map[string]any) {
			value["compromised"].(map[string]any)["path"] = "artifacts/compromised.bin"
		},
		"unknown": func(value map[string]any) { value["extra"] = true },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := iamPasswordBlocklistContractDocument()
			mutate(value)
			if err := schema.Validate(value); err == nil {
				t.Fatal("invalid blocklist coordinates accepted")
			}
		})
	}
}

func TestIAMAPIProfileRequiresPasswordBlocklistSource(t *testing.T) {
	schema := compileSchema(t, "core-api-config.schema.json")
	missing := profileSnapshot("api")
	delete(missing["variables"].(map[string]any), "HARPY_IAM_PASSWORD_BLOCKLISTS")
	if err := schema.Validate(missing); err == nil {
		t.Fatal("API profile accepted missing password blocklists")
	}
	both := profileSnapshot("api")
	both["variables"].(map[string]any)["HARPY_IAM_PASSWORD_BLOCKLISTS_FILE"] =
		"/secrets/password-blocklists.json"
	if err := schema.Validate(both); err == nil {
		t.Fatal("API profile accepted both blocklist config sources")
	}
}

func iamPasswordBlocklistContractDocument() map[string]any {
	return map[string]any{
		"version": 1, "normalization": "unicode_nfc_v1", "digest": "sha256",
		"common": map[string]any{
			"path": "/artifacts/common-v1.bin", "sha256": strings.Repeat("1", 64),
		},
		"compromised": map[string]any{
			"path": "/artifacts/compromised-v1.bin", "sha256": strings.Repeat("2", 64),
		},
	}
}
