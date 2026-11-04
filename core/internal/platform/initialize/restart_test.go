package initialize

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRestartApplicationDoesNotHaveAGooseHook(t *testing.T) {
	started := false
	if err := RestartApplication(context.Background(), func(context.Context) error {
		started = true
		return nil
	}); err != nil || !started {
		t.Fatal(err)
	}
	err := RestartApplication(context.Background(), func(context.Context) error {
		return os.ErrClosed
	})
	if err == nil || !strings.Contains(err.Error(), "stage=application") ||
		strings.Contains(err.Error(), "goose") {
		t.Fatal(err)
	}
}

func TestAPIAndWorkerMainsDoNotImportLifecycle(t *testing.T) {
	root := moduleRoot(t)
	for _, relative := range []string{"core/cmd/api/main.go", "core/cmd/worker/main.go"} {
		content, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatal(err)
		}
		source := string(content)
		for _, needle := range []string{"platform/initialize", "platform/foundation", "pressly/goose"} {
			if strings.Contains(source, needle) {
				t.Fatalf("%s imports %s", relative, needle)
			}
		}
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("module root not found")
		}
		dir = parent
	}
}
