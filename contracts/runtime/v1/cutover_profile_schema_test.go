package runtimev1

import "testing"

func TestCutoverProfileSchemaAcceptsExactOneShotSnapshots(t *testing.T) {
	for _, environment := range []string{"development", "production", "staging", "test"} {
		t.Run(environment, func(t *testing.T) {
			snapshot := cutoverSchemaSnapshot(environment)
			if err := compileSchema(t, "core-migrator-config.schema.json").Validate(snapshot); err != nil {
				t.Fatalf("valid cutover profile rejected: %v", err)
			}
		})
	}
}

func TestCutoverProfileSchemaRejectsForeignAuthorityAndSurface(t *testing.T) {
	for _, key := range []string{
		"HARPY_POSTGRES_DSN",
		"HARPY_ATTESTATION_SIGNING_KEY",
		"HARPY_RELEASE_MANIFEST_PATH",
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH",
		"HARPY_MANAGEMENT_LISTEN_ADDR",
		"HARPY_PUBLIC_LISTEN_ADDR",
	} {
		t.Run(key, func(t *testing.T) {
			snapshot := cutoverSchemaSnapshot("development")
			variables := snapshot["variables"].(map[string]any)
			variables[key] = cutoverForeignValue(key)
			if err := compileSchema(t, "core-migrator-config.schema.json").Validate(snapshot); err == nil {
				t.Fatalf("cutover profile accepted foreign key %s", key)
			}
		})
	}
}

func TestCutoverProfileSchemaRejectsPrincipalTLSAndXORDrift(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"wrong source principal", func(v map[string]any) {
			v["HARPY_CUTOVER_SOURCE_DSN"] = "postgresql://harpy_adopter@example.invalid/app?sslmode=verify-full"
		}},
		{"wrong target principal", func(v map[string]any) {
			v["HARPY_CUTOVER_TARGET_DSN"] = "postgresql://harpy_migrator@example.invalid/app?sslmode=verify-full"
		}},
		{"weak production TLS", func(v map[string]any) {
			v["HARPY_CUTOVER_SOURCE_DSN"] = "postgresql://harpy_cutover_decoder@example.invalid/app?sslmode=disable"
		}},
		{"both source DSNs", func(v map[string]any) {
			v["HARPY_CUTOVER_SOURCE_DSN_FILE"] = "/secrets/source-dsn"
		}},
		{"missing batch", func(v map[string]any) {
			delete(v, "HARPY_CUTOVER_BATCH_PATH")
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := cutoverSchemaSnapshot("production")
			test.mutate(snapshot["variables"].(map[string]any))
			if err := compileSchema(t, "core-migrator-config.schema.json").Validate(snapshot); err == nil {
				t.Fatal("unsafe cutover profile was accepted")
			}
		})
	}
}

func cutoverSchemaSnapshot(environment string) map[string]any {
	sslMode := "disable"
	if environment == "production" || environment == "staging" {
		sslMode = "verify-full"
	}
	variables := map[string]any{
		"HARPY_ENVIRONMENT":                       environment,
		"HARPY_POSTGRES_ENDPOINT_KIND":            "direct_postgres",
		"HARPY_CUTOVER_SOURCE_DSN":                "postgresql://harpy_cutover_decoder@example.invalid/app?sslmode=" + sslMode,
		"HARPY_CUTOVER_TARGET_DSN":                "postgresql://harpy_adopter@example.invalid/app?sslmode=" + sslMode,
		"HARPY_CUTOVER_LEASE_MANIFEST_PATH":       "/cutover/lease.json",
		"HARPY_CUTOVER_LEASE_SIGNATURE_PATH":      "/cutover/lease.sig",
		"HARPY_CUTOVER_TRUST_ROOT_PATH":           "/cutover/trust.json",
		"HARPY_CUTOVER_SESSION_SEAL_RECEIPT_PATH": "/cutover/seal.json",
		"HARPY_CUTOVER_BATCH_PATH":                "/cutover/batch.json",
	}
	if environment == "production" || environment == "staging" {
		variables["HARPY_OTLP_ENDPOINT"] = "https://collector.invalid:4318"
	}
	return map[string]any{"profile": "cutover_adoption", "variables": variables}
}

func cutoverForeignValue(key string) string {
	switch key {
	case "HARPY_MANAGEMENT_LISTEN_ADDR", "HARPY_PUBLIC_LISTEN_ADDR":
		return "127.0.0.1:9000"
	case "HARPY_POSTGRES_DSN":
		return "postgresql://harpy_migrator@example.invalid/app?sslmode=disable"
	default:
		return "/forbidden"
	}
}
