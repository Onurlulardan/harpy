package runtimev1

import "testing"

func TestPhase14ConfigCatalogSatisfiesFragmentSchema(t *testing.T) {
	schema := compileSchema(t, "phase14-config.schema.json")
	if err := schema.Validate(readJSONValue(t, "phase14-config.json")); err != nil {
		t.Fatalf("phase14 config catalog does not satisfy its schema: %v", err)
	}
}

func TestPhase14ProfileSchemasAcceptExactSnapshots(t *testing.T) {
	tests := []struct {
		schema  string
		profile string
	}{
		{"core-operator-config.schema.json", "operator_initialize"},
		{"core-operator-config.schema.json", "operator_identity_reconcile"},
		{"core-operator-config.schema.json", "operator_license_reconcile"},
		{"core-operator-config.schema.json", "operator_post_restore_reactivation_observer"},
		{"core-migrator-config.schema.json", "tenant_migration_executor"},
		{"core-migrator-config.schema.json", "database_backup_executor"},
		{"core-migrator-config.schema.json", "database_restore"},
	}
	for _, test := range tests {
		t.Run(test.profile, func(t *testing.T) {
			snapshot := phase14ProfileSnapshot(test.profile)
			if err := compileSchema(t, test.schema).Validate(snapshot); err != nil {
				t.Fatalf("valid phase14 profile rejected: %v", err)
			}
		})
	}
}

func TestPhase14ProfileSchemasRejectModeAndSourceLeakage(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		snapshot map[string]any
	}{
		{
			"saas api recovery key", "core-api-config.schema.json",
			snapshotWith("api", map[string]any{
				"HARPY_RECOVERY_CONTROL_SIGNING_KEY_FILE": "/recovery/control.key",
			}),
		},
		{
			"worker both license sources", "core-worker-config.schema.json",
			snapshotWith("worker", map[string]any{
				"HARPY_LICENSE_CREDENTIAL_FILE": "/license/credential",
				"HARPY_LICENSE_ISSUER_URL":      "https://issuer.invalid",
			}),
		},
		{
			"initialize without license source", "core-operator-config.schema.json",
			withoutSnapshotKey(phase14ProfileSnapshot("operator_initialize"), "HARPY_LICENSE_SNAPSHOT_FILE"),
		},
		{
			"identity with license source", "core-operator-config.schema.json",
			withPhase14Variable(phase14ProfileSnapshot("operator_identity_reconcile"),
				"HARPY_LICENSE_SNAPSHOT_FILE", "/license/snapshot.json"),
		},
		{
			"restore branches combined", "core-migrator-config.schema.json",
			withPhase14Variable(phase14ProfileSnapshot("database_restore"),
				"HARPY_RECOVERY_KIT_FILE", "/recovery/kit"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := compileSchema(t, test.schema).Validate(test.snapshot); err == nil {
				t.Fatal("invalid phase14 snapshot was accepted")
			}
		})
	}
}

func phase14ProfileSnapshot(profile string) map[string]any {
	variables := phase14StandardRuntime()
	switch profile {
	case "operator_initialize":
		variables["HARPY_INSTALLATION_MANIFEST_PATH"] = "/installation/manifest.json"
		variables["HARPY_INITIAL_ADMIN_PASSWORD_FILE"] = "/installation/password"
		variables["HARPY_IAM_LOGIN_DIRECTORY_KEYRING_FILE"] = "/iam/directory-keyring.json"
		variables["HARPY_RECOVERY_SIGNING_DIRECTORY"] = "/recovery/signing"
		variables["HARPY_LICENSE_TRUST_BUNDLE_PATH"] = "/license/trust.json"
		variables["HARPY_LICENSE_SNAPSHOT_FILE"] = "/license/snapshot.json"
	case "operator_identity_reconcile":
		addIdentityReconcileSnapshot(variables)
	case "operator_license_reconcile":
		variables["HARPY_RESTORE_RECEIPT_PATH"] = "/restore/receipt.json"
		variables["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] = "/recovery/trust.json"
		variables["HARPY_LICENSE_TRUST_BUNDLE_PATH"] = "/license/trust.json"
		variables["HARPY_LICENSE_SNAPSHOT_FILE"] = "/license/snapshot.json"
	case "operator_post_restore_reactivation_observer":
		variables = phase14DedicatedRuntime()
		addReleaseInputsSnapshot(variables)
		variables["HARPY_POST_RESTORE_REACTIVATION_POSTGRES_DSN_FILE"] =
			"/postgres/reactivation-observer-dsn"
		variables["HARPY_POST_RESTORE_REACTIVATION_INPUT_FILE"] =
			"/restore/reactivation-await.json"
	case "tenant_migration_executor":
		addReleaseInputsSnapshot(variables)
	case "database_backup_executor":
		variables = phase14DedicatedRuntime()
		addReleaseInputsSnapshot(variables)
		variables["HARPY_BACKUP_POSTGRES_DSN_FILE"] = "/postgres/backup-dsn"
		variables["HARPY_RECOVERY_STORAGE_MANIFEST_PATH"] = "/recovery/storage.json"
		variables["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] = "/recovery/trust.json"
		variables["HARPY_RECOVERY_MANIFEST_SIGNING_KEY_FILE"] = "/recovery/manifest.key"
	case "database_restore":
		variables = phase14DedicatedRuntime()
		addReleaseInputsSnapshot(variables)
		addDatabaseRestoreSnapshot(variables)
	}
	return map[string]any{"profile": profile, "variables": variables}
}

func phase14StandardRuntime() map[string]any {
	return map[string]any{
		"HARPY_ENVIRONMENT":            "development",
		"HARPY_POSTGRES_ENDPOINT_KIND": "direct_postgres",
		"HARPY_POSTGRES_DSN":           "postgresql://user@example.invalid/harpy?sslmode=disable",
	}
}

func phase14DedicatedRuntime() map[string]any {
	return map[string]any{
		"HARPY_ENVIRONMENT": "development", "HARPY_POSTGRES_ENDPOINT_KIND": "direct_postgres",
	}
}

func addIdentityReconcileSnapshot(variables map[string]any) {
	variables["HARPY_RESTORE_RECEIPT_PATH"] = "/restore/receipt.json"
	variables["HARPY_RESTORE_SECURITY_EVIDENCE_PATH"] = "/restore/security.json"
	variables["HARPY_IAM_KEY_CUSTODY_MANIFEST_PATH"] = "/iam/custody.json"
	variables["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] = "/recovery/trust.json"
	variables["HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH"] = "/authority/ack-trust.json"
}

func addDatabaseRestoreSnapshot(variables map[string]any) {
	variables["HARPY_RESTORE_CONTROL_POSTGRES_DSN_FILE"] = "/postgres/control-dsn"
	variables["HARPY_RESTORE_SCRATCH_POSTGRES_DSN_FILE"] = "/postgres/scratch-dsn"
	variables["HARPY_RECOVERY_STORAGE_MANIFEST_PATH"] = "/recovery/storage.json"
	variables["HARPY_RECOVERY_TRUST_BUNDLE_PATH"] = "/recovery/trust.json"
	variables["HARPY_RESTORE_ORCHESTRATOR_TRUST_BUNDLE_PATH"] = "/restore/orchestrator-trust.json"
	variables["HARPY_RESTORE_PROGRESS_PATH"] = "/restore/progress.json"
	variables["HARPY_RESTORE_DRAIN_ACK_PATH"] = "/restore/drain-ack.json"
	variables["HARPY_RESTORE_SWITCH_ACK_PATH"] = "/restore/switch-ack.json"
	variables["HARPY_RESTORE_RECEIPT_PATH"] = "/restore/receipt.json"
	variables["HARPY_RESTORE_PLAN_PATH"] = "/restore/plan.json"
	variables["HARPY_DR_KEK_KEYRING_FILE"] = "/recovery/dr-kek"
	variables["HARPY_RESTORE_RECEIPT_SIGNING_KEY_FILE"] = "/recovery/receipt.key"
}

func withPhase14Variable(snapshot map[string]any, key string, value any) map[string]any {
	snapshot["variables"].(map[string]any)[key] = value
	return snapshot
}

func withoutSnapshotKey(snapshot map[string]any, key string) map[string]any {
	delete(snapshot["variables"].(map[string]any), key)
	return snapshot
}
