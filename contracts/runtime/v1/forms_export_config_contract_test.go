package runtimev1

import "testing"

func TestFormsExportFCARTKeyringSchema(t *testing.T) {
	valid := map[string]any{
		"version": 1,
		"current": map[string]any{
			"kid": "forms-export-current",
			"key": "R0dHR0dHR0dHR0dHR0dHR0dHR0dHR0dHR0dHR0dHR0c",
		},
		"retained": []any{},
	}
	schema := compileSchema(t, "forms-export-fcart-kek-keyring.schema.json")
	if err := schema.Validate(valid); err != nil {
		t.Fatalf("valid Forms export keyring rejected: %v", err)
	}
	invalid := cloneJSONMap(valid)
	invalid["unknown"] = true
	if err := schema.Validate(invalid); err == nil {
		t.Fatal("keyring schema accepted unknown member")
	}
	eight := make([]any, 8)
	for index := range eight {
		eight[index] = map[string]any{
			"kid": "retained-key-" + string(rune('a'+index)),
			"key": "SEhISEhISEhISEhISEhISEhISEhISEhISEhISEhISEg",
		}
	}
	invalid = cloneJSONMap(valid)
	invalid["retained"] = eight
	if err := schema.Validate(invalid); err == nil {
		t.Fatal("keyring schema accepted eight retained keys")
	}
}

func TestFormsExportProfileCredentialRules(t *testing.T) {
	schema := compileSchema(t, "core-api-config.schema.json")
	static := snapshotWith("api", map[string]any{
		"HARPY_FORMS_EXPORT_S3_CREDENTIAL_SOURCE":      "static_files",
		"HARPY_FORMS_EXPORT_S3_ACCESS_KEY_ID_FILE":     "/secrets/access",
		"HARPY_FORMS_EXPORT_S3_SECRET_ACCESS_KEY_FILE": "/secrets/secret",
	})
	if err := schema.Validate(static); err != nil {
		t.Fatalf("static _FILE profile rejected: %v", err)
	}
	missing := snapshotWith("api", map[string]any{
		"HARPY_FORMS_EXPORT_S3_CREDENTIAL_SOURCE": "static_files",
	})
	if err := schema.Validate(missing); err == nil {
		t.Fatal("static source without files accepted")
	}
	ambientWithFile := snapshotWith("api", map[string]any{
		"HARPY_FORMS_EXPORT_S3_ACCESS_KEY_ID_FILE": "/secrets/access",
	})
	if err := schema.Validate(ambientWithFile); err == nil {
		t.Fatal("workload identity with static file accepted")
	}
}

func cloneJSONMap(source map[string]any) map[string]any {
	clone := make(map[string]any, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}
