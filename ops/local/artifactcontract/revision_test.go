package artifactcontract

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentRevisionIsTheSharedContract(t *testing.T) {
	if CurrentRevision != 13 {
		t.Fatalf("CurrentRevision = %d, want 13", CurrentRevision)
	}
	root := moduleRoot(t)
	output := filepath.Join(t.TempDir(), "secrets")
	runLocalconfig(t, root, output)
	manifest := readManifestRevision(t, filepath.Join(output, "manifest.json"))
	if manifest != CurrentRevision {
		t.Fatalf("localconfig manifest revision = %d, want %d", manifest, CurrentRevision)
	}
	assertProjectorRevision(t, root, output)
}

func assertProjectorRevision(t *testing.T, root, source string) {
	t.Helper()
	accepted := runProjector(t, root, source, filepath.Join(t.TempDir(), "targets"))
	if strings.Contains(accepted, "source manifest identity is invalid") {
		t.Fatalf("projector rejected CurrentRevision: %s", accepted)
	}
	stale := filepath.Join(t.TempDir(), "stale-secrets")
	copySource(t, source, stale)
	replaceRevision(t, filepath.Join(stale, "manifest.json"), CurrentRevision, CurrentRevision-1)
	rejected := runProjector(t, root, stale, filepath.Join(t.TempDir(), "stale-targets"))
	if !strings.Contains(rejected, "source manifest identity is invalid") {
		t.Fatalf("projector accepted revision %d: %s", CurrentRevision-1, rejected)
	}
}

func runLocalconfig(t *testing.T, root, output string) {
	t.Helper()
	envLocal := filepath.Join(root, ".env.local")
	_, statErr := os.Lstat(envLocal)
	createdEnv := os.IsNotExist(statErr)
	command := exec.Command("go", "run", "./tools/localconfig", "-output", output)
	command.Dir = root
	combined, err := command.CombinedOutput()
	if createdEnv {
		if removeErr := os.Remove(envLocal); removeErr != nil && !os.IsNotExist(removeErr) {
			t.Fatal(removeErr)
		}
	}
	if err != nil {
		t.Fatalf("localconfig: %v\n%s", err, combined)
	}
}

func runProjector(t *testing.T, root, source, targets string) string {
	t.Helper()
	if err := os.MkdirAll(targets, 0o700); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(
		"go", "run", "./ops/local/cmd/secret-projector",
		"-source", source, "-targets", targets,
	)
	command.Dir = root
	combined, _ := command.CombinedOutput()
	return string(combined)
}

func readManifestRevision(t *testing.T, path string) int {
	t.Helper()
	var manifest struct {
		ArtifactContractRevision int `json:"artifact_contract_revision"`
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatal(err)
	}
	return manifest.ArtifactContractRevision
}

func replaceRevision(t *testing.T, path string, from, to int) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	previous := fmt.Sprintf(`"artifact_contract_revision":%d`, from)
	updated := strings.ReplaceAll(string(content), previous,
		fmt.Sprintf(`"artifact_contract_revision":%d`, to))
	if updated == string(content) {
		t.Fatalf("manifest revision %d was not present", from)
	}
	if err := os.WriteFile(path, []byte(updated), 0o600); err != nil {
		t.Fatal(err)
	}
}

func copySource(t *testing.T, source, destination string) {
	t.Helper()
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("source contains directory %s", entry.Name())
		}
		content, err := os.ReadFile(filepath.Join(source, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(
			filepath.Join(destination, entry.Name()), content, 0o600,
		); err != nil {
			t.Fatal(err)
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(directory, "go.mod")); statErr == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("module root is unavailable")
		}
		directory = parent
	}
}
