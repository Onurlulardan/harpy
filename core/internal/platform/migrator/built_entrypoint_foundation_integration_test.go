//go:build integration

package migrator

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func migrateBuiltEntrypointDatabase(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
	stage builtEntrypointStage,
	migratorBinary string,
) builtEntrypointDeployEvidence {
	t.Helper()
	evidence := prepareBuiltEntrypointDeployEvidence(t, ctx, database, stage)
	environment := builtEntrypointDeployEnvironment(database, evidence)
	up := runBuiltDeployCommand(t, ctx, migratorBinary, environment, "up")
	if up.Up == nil || len(up.Up.AppliedVersions) == 0 {
		t.Fatalf("built migrator deploy up applied no migration: %#v", up)
	}
	verified := runBuiltDeployCommand(t, ctx, migratorBinary, environment, "verify")
	if verified.Verify == nil || verified.Verify.VerifiedRunID != up.Up.RunID ||
		verified.Verify.DatabaseGeneration < 1 ||
		verified.Verify.MinimumRuntimeGeneration < 1 {
		t.Fatalf("built migrator verify differs from up: %#v %#v", up, verified)
	}
	assertBuiltDeployFinalWrite(t, ctx, database, up.Up.RunID, evidence)
	grantBuiltEntrypointRuntime(t, ctx, database)
	assertBuiltEntrypointPrivilegeDenials(t, ctx, database)
	return evidence
}

func runBuiltDeployCommand(
	t *testing.T,
	ctx context.Context,
	binary string,
	environment []string,
	action string,
) runtimev1.MigratorCommandResult {
	t.Helper()
	command := exec.CommandContext(ctx, binary, "deploy", action)
	command.Env = config.IntegrationSubprocessEnvironment(environment)
	var stdout bytes.Buffer
	command.Stdout = &stdout
	logFile, err := os.CreateTemp(t.TempDir(), "deploy-"+action+"-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()
	command.Stderr = logFile
	if err := command.Run(); err != nil {
		t.Fatalf(
			"built migrator deploy %s failed: %v: %s",
			action, err, readBuiltEntrypointLog(logFile),
		)
	}
	var result runtimev1.MigratorCommandResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode built migrator deploy %s result: %v: %s", action, err, stdout.Bytes())
	}
	if err := result.Validate(); err != nil || result.Command != action {
		t.Fatalf("invalid built migrator deploy %s result: %#v err=%v", action, result, err)
	}
	return result
}

func grantBuiltEntrypointRuntime(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
) {
	t.Helper()
	statements := []string{
		"GRANT USAGE ON SCHEMA platform TO harpy_runtime",
		"GRANT SELECT ON platform.environment_identity, platform.bootstrap_receipts, " +
			"platform.extension_registry, platform.migration_runs, platform.schema_contracts, " +
			"platform.database_generation_manifests, platform.release_generations " +
			"TO harpy_runtime",
	}
	for _, statement := range statements {
		if _, err := database.admin.Exec(ctx, statement); err != nil {
			t.Fatalf("grant built entrypoint runtime contract: %v", err)
		}
	}
}
