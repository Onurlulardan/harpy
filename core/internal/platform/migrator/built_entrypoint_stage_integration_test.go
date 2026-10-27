//go:build integration

package migrator

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type builtEntrypointStage struct {
	root            string
	directory       string
	migrationDigest string
	extensionDigest string
}

func prepareBuiltEntrypointStage(
	t *testing.T,
	ctx context.Context,
) builtEntrypointStage {
	t.Helper()
	root := builtEntrypointRepositoryRoot(t)
	directory := filepath.Join(t.TempDir(), "backend-v2")
	copyBuiltEntrypointSources(t, root, directory)
	migrationDigest := generateCurrentMigrationManifest(t, ctx, root, directory)
	extensionDigest := builtEntrypointExtensionDigest(t)
	contract := runtimev1.BinarySchemaContract{
		ContractVersion: runtimev1.BinarySchemaContractVersion,
		Component:       "core", RuntimeGeneration: 1, MinimumDatabaseGeneration: 1,
		MigrationManifestSHA256: migrationDigest,
		ExtensionManifestSHA256: extensionDigest,
		PostgresImageDigest:     runtimev1.PostgresImageDigest,
	}
	document, err := attestation.MarshalCanonical(contract)
	if err != nil {
		t.Fatalf("encode canonical built entrypoint contract: %v", err)
	}
	target := filepath.Join(
		directory, "contracts", "runtime", "v1", "binary-schema-contract.json",
	)
	if err := os.WriteFile(target, document, 0o600); err != nil {
		t.Fatalf("write staged built entrypoint contract: %v", err)
	}
	return builtEntrypointStage{
		root: root, directory: directory,
		migrationDigest: migrationDigest, extensionDigest: extensionDigest,
	}
}

func builtEntrypointRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve built entrypoint source path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../.."))
}

func copyBuiltEntrypointSources(t *testing.T, root, target string) {
	t.Helper()
	embedded := builtEntrypointEmbeddedFiles(t, root)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("staged source contains symlink %s", path)
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		extension := strings.ToLower(filepath.Ext(base))
		_, embeddedAsset := embedded[filepath.Clean(path)]
		copyFile := base == "go.mod" || base == "go.sum" ||
			extension == ".go" || extension == ".sql" ||
			extension == ".json" || extension == ".yaml" || embeddedAsset
		if !copyFile || strings.HasSuffix(base, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			return err
		}
		mode := os.FileMode(0o600)
		if embeddedAsset {
			mode = 0o400
		}
		return os.WriteFile(destination, content, mode)
	})
	if err != nil {
		t.Fatalf("stage built entrypoint source: %v", err)
	}
}

func builtEntrypointEmbeddedFiles(t *testing.T, root string) map[string]struct{} {
	t.Helper()
	const format = `{{range .EmbedFiles}}{{$.Dir}}{{printf "%c" 0}}{{.}}{{"\n"}}{{end}}`
	command := exec.Command("go", "list", "-f", format, "./...")
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		t.Fatalf("list staged source embed assets: %v", err)
	}
	files := map[string]struct{}{}
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		directory, name, found := bytes.Cut(line, []byte{0})
		if !found {
			continue
		}
		candidate := filepath.Clean(filepath.Join(string(directory), string(name)))
		relative, err := filepath.Rel(root, candidate)
		if err != nil || relative == ".." ||
			strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Fatalf("embed asset escapes repository: %s", name)
		}
		files[candidate] = struct{}{}
	}
	return files
}

func TestBuiltEntrypointStageCopiesEmbeddedTZDB(t *testing.T) {
	root, target := builtEntrypointRepositoryRoot(t), t.TempDir()
	copyBuiltEntrypointSources(t, root, target)
	relative := filepath.Join("core", "internal", "modules", "workflow",
		"adapters", "tzdb", "assets", "2026c.zip")
	source, sourceErr := os.ReadFile(filepath.Join(root, relative))
	staged, stagedErr := os.ReadFile(filepath.Join(target, relative))
	if sourceErr != nil || stagedErr != nil || !bytes.Equal(source, staged) {
		t.Fatalf("staged embedded TZDB differs: source=%v staged=%v", sourceErr, stagedErr)
	}
}

func generateCurrentMigrationManifest(
	t *testing.T,
	ctx context.Context,
	root string,
	stage string,
) string {
	t.Helper()
	command := exec.CommandContext(
		ctx, "go", "run", "./tools/migrationmanifest", "-root", stage, "-write",
	)
	command.Dir = stage
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("generate current staged migration manifest: %v: %s", err, output)
	}
	digest := parseBuiltEntrypointManifestDigest(t, string(output))
	name := digest + ".migration-manifest.json"
	stagedPath := filepath.Join(stage, "core", "internal", "platform", "migration", "release", name)
	sourcePath := filepath.Join(root, "core", "internal", "platform", "migration", "release", name)
	staged, err := os.ReadFile(stagedPath)
	if err != nil {
		t.Fatalf("read generated migration manifest: %v", err)
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("current canonical migration manifest %s is not finalized: %v", digest, err)
	}
	if !bytes.Equal(source, staged) {
		t.Fatalf("current canonical migration manifest %s differs from source registry", digest)
	}
	return digest
}

func parseBuiltEntrypointManifestDigest(t *testing.T, output string) string {
	t.Helper()
	for _, line := range strings.Split(strings.ReplaceAll(output, "\r", ""), "\n") {
		if digest, found := strings.CutPrefix(line, "sha256: "); found && len(digest) == 64 {
			return digest
		}
	}
	t.Fatalf("migration manifest command omitted digest: %s", output)
	return ""
}
