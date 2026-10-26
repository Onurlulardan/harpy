package runtimev1

import "testing"

func TestIAMSystemTransferCredentialContracts(t *testing.T) {
	schema := compileSchema(t, "core-worker-config.schema.json")
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"export static credentials incomplete", func(value map[string]any) {
			value["HARPY_IAM_SYSTEM_EXPORT_S3_CREDENTIAL_SOURCE"] = "static_files"
		}},
		{"import static credentials incomplete", func(value map[string]any) {
			value["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_CREDENTIAL_SOURCE"] = "static_files"
		}},
		{"export workload identity with static file", func(value map[string]any) {
			value["HARPY_IAM_SYSTEM_EXPORT_S3_ACCESS_KEY_ID_FILE"] = "/secrets/access-key"
		}},
		{"import workload identity with static file", func(value map[string]any) {
			value["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_SESSION_TOKEN_FILE"] = "/secrets/session-token"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := profileSnapshot("worker")
			test.mutate(snapshot["variables"].(map[string]any))
			if err := schema.Validate(snapshot); err == nil {
				t.Fatal("invalid IAM system transfer artifact profile was accepted")
			}
		})
	}
}

func TestIAMSystemTransferProductionRequiresTLS(t *testing.T) {
	snapshot := profileSnapshot("worker")
	variables := snapshot["variables"].(map[string]any)
	variables["HARPY_ENVIRONMENT"] = "production"
	variables["HARPY_POSTGRES_DSN"] =
		"postgresql://user@example.invalid/harpy?sslmode=verify-full"
	variables["HARPY_OTLP_ENDPOINT"] = "https://otlp.example.invalid"
	variables["HARPY_RABBITMQ_URL"] = "amqps://worker:secret@rabbit.invalid/%2f"
	variables["HARPY_IAM_SYSTEM_IMPORT_SOURCE_S3_ENDPOINT"] = "http://s3.invalid"
	if err := compileSchema(t, "core-worker-config.schema.json").Validate(snapshot); err == nil {
		t.Fatal("production IAM system transfer HTTP endpoint was accepted")
	}
}
