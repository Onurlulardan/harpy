package runtimev1

import "testing"

func TestCutoverFinalizationProfileSchemaAcceptsExactSnapshot(t *testing.T) {
	for _, environment := range []string{"development", "production", "staging", "test"} {
		snapshot := cutoverFinalizationSchemaSnapshot(environment)
		if err := compileSchema(t, "core-migrator-config.schema.json").Validate(snapshot); err != nil {
			t.Fatalf("valid finalization profile rejected: %v", err)
		}
	}
}

func TestCutoverFinalizationProfileRejectsForeignAuthority(t *testing.T) {
	for _, key := range []string{
		"HARPY_CUTOVER_SOURCE_DSN", "HARPY_CUTOVER_TARGET_DSN",
		"HARPY_CUTOVER_BATCH_PATH", "HARPY_ATTESTATION_SIGNING_KEY",
		"HARPY_MANAGEMENT_LISTEN_ADDR",
	} {
		snapshot := cutoverFinalizationSchemaSnapshot("development")
		snapshot["variables"].(map[string]any)[key] = "/forbidden"
		if err := compileSchema(t, "core-migrator-config.schema.json").Validate(snapshot); err == nil {
			t.Fatalf("finalization profile accepted foreign key %s", key)
		}
	}
}

func TestCutoverFinalizationProfileRejectsPrincipalTLSAndEvidenceDrift(t *testing.T) {
	tests := []func(map[string]any){
		func(v map[string]any) {
			v["HARPY_CUTOVER_FINALIZER_DSN"] = "postgresql://harpy_adopter@example.invalid/app?sslmode=verify-full"
		},
		func(v map[string]any) {
			v["HARPY_CUTOVER_FINALIZER_DSN"] = "postgresql://harpy_cutover_finalizer@example.invalid/app?sslmode=disable"
		},
		func(v map[string]any) { delete(v, "HARPY_CUTOVER_CREDENTIAL_CLEANUP_RECEIPT_PATH") },
		func(v map[string]any) { v["HARPY_CUTOVER_FINALIZER_DSN_FILE"] = "/secret/dsn" },
	}
	for _, mutate := range tests {
		snapshot := cutoverFinalizationSchemaSnapshot("production")
		mutate(snapshot["variables"].(map[string]any))
		if err := compileSchema(t, "core-migrator-config.schema.json").Validate(snapshot); err == nil {
			t.Fatal("unsafe finalization profile was accepted")
		}
	}
}

func cutoverFinalizationSchemaSnapshot(environment string) map[string]any {
	sslMode := "disable"
	if environment == "production" || environment == "staging" {
		sslMode = "verify-full"
	}
	variables := map[string]any{
		"HARPY_ENVIRONMENT":                             environment,
		"HARPY_POSTGRES_ENDPOINT_KIND":                  "direct_postgres",
		"HARPY_CUTOVER_FINALIZER_DSN":                   "postgresql://harpy_cutover_finalizer@example.invalid/app?sslmode=" + sslMode,
		"HARPY_CUTOVER_LEASE_MANIFEST_PATH":             "/cutover/lease.json",
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH":            "/cutover/lease.sig",
		"HARPY_CUTOVER_TRUST_ROOT_PATH":                 "/cutover/trust.json",
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH":       "/cutover/seal.json",
		"HARPY_CUTOVER_CREDENTIAL_CLEANUP_RECEIPT_PATH": "/cutover/cleanup.json",
		"HARPY_CUTOVER_FINALIZER_LEASE_MANIFEST_PATH":   "/cutover/finalizer.json",
		"HARPY_CUTOVER_FINALIZER_LEASE_SIGNATURE_PATH":  "/cutover/finalizer.sig",
	}
	if environment == "production" || environment == "staging" {
		variables["HARPY_OTLP_ENDPOINT"] = "https://collector.invalid:4318"
	}
	return map[string]any{"profile": "cutover_finalization", "variables": variables}
}
