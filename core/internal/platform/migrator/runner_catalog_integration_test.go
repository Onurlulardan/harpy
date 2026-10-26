//go:build integration

package migrator

import (
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"testing/fstest"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func runnerIntegrationCatalog(
	t *testing.T,
) (fstest.MapFS, []postgresv1.SelectedMigration) {
	return runnerIntegrationCatalogForOrigin(t, postgresv1.DatabaseOriginFresh)
}

func runnerIntegrationCatalogForOrigin(
	t *testing.T,
	origin postgresv1.DatabaseOrigin,
) (fstest.MapFS, []postgresv1.SelectedMigration) {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve runner integration source path")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../.."))
	contents, selected := currentRunnerIntegrationPlatformCatalog(
		t, repositoryRoot, origin,
	)
	files := fstest.MapFS{}
	for index := range selected {
		candidate := &selected[index]
		basename := path.Base(candidate.Path)
		if _, duplicate := files[basename]; duplicate {
			t.Fatalf("integration migration basename %s is duplicated", basename)
		}
		files[basename] = &fstest.MapFile{
			Data: append([]byte(nil), contents[candidate.Path]...),
			Mode: 0o444,
		}
		candidate.Phase = "foundation"
		candidate.TargetDatabaseGeneration = 1
		candidate.NextMinimumRuntimeGeneration = 1
		if candidate.Version >= runnerIntegrationExpandVersion {
			candidate.Phase = "expand"
			candidate.TargetDatabaseGeneration = 2
		}
	}
	if err := validatePlatformSelection(selected); err != nil {
		t.Fatalf("validate integration migration selection: %v", err)
	}
	return files, selected
}

const runnerIntegrationExpandVersion = int64(20260714010400)

func runnerIntegrationBaselineCatalog(
	t *testing.T,
) (fstest.MapFS, []postgresv1.SelectedMigration) {
	t.Helper()
	files, selected := runnerIntegrationCatalog(t)
	baselineFiles := fstest.MapFS{}
	for name, file := range files {
		version, err := strconv.ParseInt(name[:14], 10, 64)
		if err != nil {
			t.Fatalf("parse baseline migration version %s: %v", name, err)
		}
		if version < runnerIntegrationExpandVersion {
			baselineFiles[name] = file
		}
	}
	baseline := make([]postgresv1.SelectedMigration, 0, len(selected))
	for _, migration := range selected {
		if migration.Version < runnerIntegrationExpandVersion {
			baseline = append(baseline, migration)
		}
	}
	return baselineFiles, baseline
}
