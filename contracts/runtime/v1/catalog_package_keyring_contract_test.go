package runtimev1

import "testing"

func TestCatalogPackageKeyringSchemaAndProfileSourceAreExact(t *testing.T) {
	document := map[string]any{
		"version": 1,
		"keys": []any{map[string]any{
			"key_id":     "first-party-2026-01",
			"public_key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		}},
	}
	if err := compileSchema(t, "catalog-package-public-keyring.schema.json").Validate(document); err != nil {
		t.Fatalf("valid Catalog package keyring rejected: %v", err)
	}
	for _, candidate := range []struct {
		schema, profile string
	}{
		{"core-api-config.schema.json", "api"},
		{"core-worker-config.schema.json", "worker"},
		{"core-migrator-config.schema.json", "migrator_deploy"},
	} {
		schema := compileSchema(t, candidate.schema)
		missing := profileSnapshot(candidate.profile)
		delete(missing["variables"].(map[string]any),
			"HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING")
		if err := schema.Validate(missing); err == nil {
			t.Fatalf("%s accepted missing Catalog keyring", candidate.profile)
		}
		fileOnly := profileSnapshot(candidate.profile)
		variables := fileOnly["variables"].(map[string]any)
		delete(variables, "HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING")
		variables["HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING_FILE"] =
			"/secrets/catalog-package-public-keyring.json"
		if err := schema.Validate(fileOnly); err != nil {
			t.Fatalf("%s rejected file-only Catalog keyring: %v", candidate.profile, err)
		}
	}
}

func validCatalogPackageKeyringContractValue() string {
	return `{"version":1,"keys":[{"key_id":"first-party-2026-01","public_key":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}]}`
}
