package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseMaterialLoadsMigratorDeploy(t *testing.T) {
	root := t.TempDir()
	lines, err := releaseMaterial(root, "single-tenant")
	if err != nil {
		t.Fatalf("release material: %v", err)
	}
	base := []string{
		"HARPY_ENVIRONMENT=development",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN=postgres://harpy:secret@127.0.0.1:5432/harpy?sslmode=disable",
		"HARPY_LOG_LEVEL=info",
	}
	if err = loadMigratorDeploy(t, append(base, lines...)); err != nil {
		t.Fatalf("LoadMigratorDeploy: %v", err)
	}
	directory := filepath.Join(root, "deploy", "single-tenant", "secrets", "release")
	prefix := directory + string(os.PathSeparator)
	for _, line := range lines {
		key, path, found := strings.Cut(line, "=")
		info, statErr := os.Stat(path)
		if !found || !filepath.IsAbs(path) || !strings.HasPrefix(path, prefix) ||
			statErr != nil || info.Size() == 0 || !info.Mode().IsRegular() {
			t.Fatalf("%s was not written: %v", key, statErr)
		}
		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode = %o", key, info.Mode().Perm())
		}
	}
}

func TestReleaseMaterialRejectsUnknownProfile(t *testing.T) {
	if _, err := releaseMaterial(t.TempDir(), "saas"); err == nil {
		t.Fatal("unknown profile was accepted")
	}
}

func loadMigratorDeploy(t *testing.T, lines []string) error {
	t.Helper()
	_, caller, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("test file path is unavailable")
	}
	module := filepath.Clean(filepath.Join(filepath.Dir(caller), "..", ".."))
	directory := filepath.Join(
		module, "core", "internal", "platform", "config", "testdata", "localmigrator",
	)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	t.Cleanup(func() { _ = os.RemoveAll(directory) })
	source := filepath.Join(directory, "main.go")
	if err := os.WriteFile(source, []byte(migratorDeployLoader), 0o600); err != nil {
		return err
	}
	envFile := filepath.Join(t.TempDir(), "migrator.env")
	if err := os.WriteFile(envFile, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		return err
	}
	command := exec.Command(goTool(), "run", "-mod=readonly", ".", envFile)
	command.Dir = directory
	command.Env = append(os.Environ(), "GOPROXY=off", "GOTOOLCHAIN=local", "GOFLAGS=-mod=readonly")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	return nil
}

const migratorDeployLoader = `package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func main() {
	os.Clearenv()
	file, err := os.Open(os.Args[1])
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || key == "" || value == "" {
			os.Stderr.WriteString("malformed env line")
			os.Exit(1)
		}
		if err = os.Setenv(key, value); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
	}
	if err = scanner.Err(); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	if _, err = config.LoadMigratorDeploy(); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}
`
