package billingv1

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBindRoutesPublishesEveryExplicitOperation(t *testing.T) {
	bindings, err := new(Handler).BindRoutes()
	if err != nil || len(bindings) != 26 {
		t.Fatalf("bind routes: count=%d err=%v", len(bindings), err)
	}
	seen := make(map[string]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding.Handler == nil {
			t.Fatalf("nil handler for %s %s", binding.Method, binding.Pattern)
		}
		key := binding.Method + " " + binding.Pattern
		if _, duplicate := seen[key]; duplicate {
			t.Fatalf("duplicate compatibility binding %s", key)
		}
		seen[key] = struct{}{}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("repository root was not found")
		}
		directory = parent
	}
}
