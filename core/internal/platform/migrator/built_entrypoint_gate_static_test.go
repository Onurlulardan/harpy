package migrator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBuiltEntrypointGateUsesBuiltDeployUpAndVerify(t *testing.T) {
	document := readBuiltEntrypointGateSource(t, "built_entrypoint_foundation_integration_test.go")
	required := []string{
		`runBuiltDeployCommand(t, ctx, migratorBinary, environment, "up")`,
		`runBuiltDeployCommand(t, ctx, migratorBinary, environment, "verify")`,
		"assertBuiltDeployFinalWrite",
	}
	for _, fragment := range required {
		if !strings.Contains(document, fragment) {
			t.Fatalf("built entrypoint gate omitted %q", fragment)
		}
	}
	forbidden := []string{"newRunnerIntegrationFixture", "runner.Up(", "runner.Verify("}
	for _, fragment := range forbidden {
		if strings.Contains(document, fragment) {
			t.Fatalf("built entrypoint gate retained in-process shortcut %q", fragment)
		}
	}
}

func TestBuiltEntrypointGateUsesProductionRoleFixture(t *testing.T) {
	document := readBuiltEntrypointGateSource(t, "production_role_fixture_integration_test.go")
	for _, role := range []string{
		"harpy_cluster_admin", "harpy_extension_admin",
		"harpy_adopter", "harpy_cutover_decoder", "harpy_cutover_finalizer",
	} {
		if !strings.Contains(document, role) {
			t.Fatalf("production role fixture omitted %s", role)
		}
	}
}

func readBuiltEntrypointGateSource(t *testing.T, name string) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve built entrypoint static test path")
	}
	document, err := os.ReadFile(filepath.Join(filepath.Dir(source), name))
	if err != nil {
		t.Fatal(err)
	}
	return string(document)
}
