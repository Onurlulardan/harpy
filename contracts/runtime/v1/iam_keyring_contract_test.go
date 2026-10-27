package runtimev1

import "testing"

func TestIAMKeyringSchemasAcceptExactWire(t *testing.T) {
	access := map[string]any{
		"version": 1,
		"active": map[string]any{
			"kid": iamContractKID(10), "seed": iamContractMaterial(10),
			"activated_at": "2026-07-01T00:00:00Z", "sign_until": "2026-07-31T00:00:00Z",
		},
		"retired_public": []any{},
	}
	if err := compileSchema(t, "iam-access-token-keyring.schema.json").Validate(access); err != nil {
		t.Fatalf("valid access keyring rejected: %v", err)
	}
	hmac := iamHMACContractDocument("refresh_token_v1", 20)
	if err := compileSchema(t, "iam-hmac-keyring.schema.json").Validate(hmac); err != nil {
		t.Fatalf("valid HMAC keyring rejected: %v", err)
	}
}

func TestIAMAccessSchemaRejectsStructuralDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"version", func(value map[string]any) { value["version"] = 2 }},
		{"unknown root", func(value map[string]any) { value["extra"] = true }},
		{"missing active", func(value map[string]any) { delete(value, "active") }},
		{"short kid", func(value map[string]any) { value["active"].(map[string]any)["kid"] = "short" }},
		{"padded seed", func(value map[string]any) { value["active"].(map[string]any)["seed"] = iamContractMaterial(10) + "=" }},
		{"retired private", func(value map[string]any) {
			value["retired_public"] = []any{map[string]any{
				"kid": iamContractKID(9), "public_key": iamContractMaterial(9),
				"sign_until": "2026-07-01T00:00:00Z", "publish_until": "2026-08-01T00:00:00Z",
				"seed": iamContractMaterial(9),
			}}
		}},
	}
	schema := compileSchema(t, "iam-access-token-keyring.schema.json")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := iamAccessContractDocument()
			test.mutate(value)
			if err := schema.Validate(value); err == nil {
				t.Fatal("invalid access keyring accepted")
			}
		})
	}
}

func TestIAMHMACSchemaRejectsStructuralDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"purpose", func(value map[string]any) { value["purpose"] = "unknown_v1" }},
		{"unknown", func(value map[string]any) { value["extra"] = true }},
		{"missing retained", func(value map[string]any) { delete(value, "retained") }},
		{"padded key", func(value map[string]any) { value["active"].(map[string]any)["key"] = iamContractMaterial(20) + "=" }},
		{"retained missing verify", func(value map[string]any) {
			value["retained"] = []any{map[string]any{
				"kid": iamContractKID(19), "key": iamContractMaterial(19),
				"activated_at": "2026-04-02T00:00:00Z", "issue_until": "2026-07-01T00:00:00Z",
			}}
		}},
	}
	schema := compileSchema(t, "iam-hmac-keyring.schema.json")
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := iamHMACContractDocument("refresh_token_v1", 20)
			test.mutate(value)
			if err := schema.Validate(value); err == nil {
				t.Fatal("invalid HMAC keyring accepted")
			}
		})
	}
}

func TestIAMProfileSchemasRequireExactSecretSources(t *testing.T) {
	schema := compileSchema(t, "core-api-config.schema.json")
	keys := []string{
		"HARPY_IAM_ACCESS_TOKEN_KEYRING", "HARPY_IAM_REFRESH_TOKEN_KEYRING",
		"HARPY_IAM_CSRF_KEYRING", "HARPY_IAM_LOGIN_DIRECTORY_KEYRING",
		"HARPY_IAM_AUTH_THROTTLE_KEYRING", "HARPY_IAM_SESSION_NETWORK_KEYRING",
		"HARPY_IAM_ONE_TIME_TOKEN_KEYRING", "HARPY_IAM_RECOVERY_CODE_KEYRING",
	}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			missing := profileSnapshot("api")
			delete(missing["variables"].(map[string]any), key)
			if err := schema.Validate(missing); err == nil {
				t.Fatal("missing IAM keyring accepted")
			}
			fileOnly := profileSnapshot("api")
			variables := fileOnly["variables"].(map[string]any)
			delete(variables, key)
			variables[key+"_FILE"] = "/secrets/iam-keyring.json"
			if err := schema.Validate(fileOnly); err != nil {
				t.Fatalf("file-only IAM keyring rejected: %v", err)
			}
			both := profileSnapshot("api")
			both["variables"].(map[string]any)[key+"_FILE"] = "/secrets/iam-keyring.json"
			if err := schema.Validate(both); err == nil {
				t.Fatal("both IAM keyring sources accepted")
			}
		})
	}
}

func TestIAMWorkerSchemaRejectsAPISecretsAndRawKMSKeys(t *testing.T) {
	worker := profileSnapshot("worker")
	worker["variables"].(map[string]any)["HARPY_IAM_ACCESS_TOKEN_KEYRING"] =
		iamAccessContractValue(10)
	if err := compileSchema(t, "core-worker-config.schema.json").Validate(worker); err == nil {
		t.Fatal("worker accepted access signing material")
	}
	api := profileSnapshot("api")
	api["variables"].(map[string]any)["HARPY_IAM_TOTP_KEY"] = iamContractMaterial(31)
	if err := compileSchema(t, "core-api-config.schema.json").Validate(api); err == nil {
		t.Fatal("API accepted raw TOTP key surface")
	}
}

func TestIAMAPIProfileRejectsNoncanonicalKMSKeyVersions(t *testing.T) {
	schema := compileSchema(t, "core-api-config.schema.json")
	for _, keyID := range []string{
		"kms://iam/totp/current",
		"gcp-kms://projects/harpy-prod/locations/europe-west1/" +
			"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/0",
	} {
		profile := profileSnapshot("api")
		profile["variables"].(map[string]any)["HARPY_IAM_TOTP_KMS_KEY_ID"] = keyID
		if err := schema.Validate(profile); err == nil {
			t.Fatalf("invalid KMS key version accepted: %q", keyID)
		}
	}
}

func iamAccessContractDocument() map[string]any {
	return map[string]any{
		"version": 1,
		"active": map[string]any{
			"kid": iamContractKID(10), "seed": iamContractMaterial(10),
			"activated_at": "2026-07-01T00:00:00Z", "sign_until": "2026-07-31T00:00:00Z",
		},
		"retired_public": []any{},
	}
}

func iamHMACContractDocument(purpose string, marker byte) map[string]any {
	return map[string]any{
		"version": 1, "purpose": purpose,
		"active": map[string]any{
			"kid": iamContractKID(marker), "key": iamContractMaterial(marker),
			"activated_at": "2026-07-01T00:00:00Z", "issue_until": "2026-09-29T00:00:00Z",
		},
		"retained": []any{},
	}
}
