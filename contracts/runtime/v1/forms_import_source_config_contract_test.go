package runtimev1

import "testing"

func TestFormsImportSourceFCARTKeyringSchema(t *testing.T) {
	valid := map[string]any{"version": 1,
		"current": map[string]any{"kid": "forms-import-source-current",
			"key": "W1tbW1tbW1tbW1tbW1tbW1tbW1tbW1tbW1tbW1tbW1s"},
		"retained": []any{}}
	schema := compileSchema(t, "forms-import-source-fcart-kek-keyring.schema.json")
	if err := schema.Validate(valid); err != nil {
		t.Fatalf("valid Forms import-source keyring rejected: %v", err)
	}
	invalid := cloneJSONMap(valid)
	invalid["unknown"] = true
	if err := schema.Validate(invalid); err == nil {
		t.Fatal("import-source keyring schema accepted unknown member")
	}
}

func TestFormsImportSourceProfileCredentialAndTLSRules(t *testing.T) {
	schema := compileSchema(t, "core-api-config.schema.json")
	static := snapshotWith("api", map[string]any{
		"HARPY_FORMS_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE":      "static_files",
		"HARPY_FORMS_IMPORT_SOURCE_S3_ACCESS_KEY_ID_FILE":     "/secrets/access",
		"HARPY_FORMS_IMPORT_SOURCE_S3_SECRET_ACCESS_KEY_FILE": "/secrets/secret"})
	if err := schema.Validate(static); err != nil {
		t.Fatalf("import-source static _FILE profile rejected: %v", err)
	}
	missing := snapshotWith("api", map[string]any{
		"HARPY_FORMS_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE": "static_files"})
	if err := schema.Validate(missing); err == nil {
		t.Fatal("import-source static credentials without files accepted")
	}
	productionHTTP := snapshotWith("api", map[string]any{
		"HARPY_ENVIRONMENT":                     "production",
		"HARPY_POSTGRES_DSN":                    "postgresql://user@example.invalid/harpy?sslmode=verify-full",
		"HARPY_OTLP_ENDPOINT":                   "https://otel.example.invalid",
		"HARPY_FORMS_IMPORT_SOURCE_S3_ENDPOINT": "http://s3.example.invalid"})
	if err := schema.Validate(productionHTTP); err == nil {
		t.Fatal("production import-source plaintext endpoint accepted")
	}
}
