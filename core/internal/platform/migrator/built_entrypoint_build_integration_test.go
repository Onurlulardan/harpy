//go:build integration

package migrator

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type builtEntrypointBinaries struct {
	api      string
	worker   string
	migrator string
}

func buildEntrypointBinaries(
	t *testing.T,
	ctx context.Context,
	stage builtEntrypointStage,
) builtEntrypointBinaries {
	t.Helper()
	directory := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	variable := "github.com/Onurlulardan/harpy/core/internal/platform/buildinfo."
	ldflags := strings.Join([]string{
		"-X", variable + "releaseVersion=2.0.0-entrypoint-gate",
		"-X", variable + "commit=" + strings.Repeat("a", 40),
		"-X", variable + "builtAt=2026-07-14T00:00:00Z",
	}, " ")
	command := exec.CommandContext(
		ctx, "go", "build", "-buildvcs=false", "-trimpath", "-ldflags", ldflags,
		"-o", directory,
		"./core/cmd/api", "./core/cmd/worker", "./core/cmd/migrator",
	)
	command.Dir = stage.directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build staged entrypoints: %v: %s", err, output)
	}
	binaries := builtEntrypointBinaries{
		api:      filepath.Join(directory, builtEntrypointBinaryName("api")),
		worker:   filepath.Join(directory, builtEntrypointBinaryName("worker")),
		migrator: filepath.Join(directory, builtEntrypointBinaryName("migrator")),
	}
	for _, path := range []string{binaries.api, binaries.worker, binaries.migrator} {
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			t.Fatalf("go build omitted entrypoint binary %s", path)
		}
	}
	return binaries
}

func builtEntrypointBinaryName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
