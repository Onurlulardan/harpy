package billingv1

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type characterizedRoute struct {
	ID        string   `json:"id"`
	Method    string   `json:"method"`
	Path      string   `json:"path"`
	Consumers []string `json:"consumers"`
}

func TestRoutesMatchExactBillingV1Characterization(t *testing.T) {
	want := readCharacterizedRoutes(t)
	got := Routes()
	if len(want) != 26 || len(got) != len(want) {
		t.Fatalf("route count want=%d got=%d", len(want), len(got))
	}
	byID := make(map[string]Route, len(got))
	keys := make(map[string]struct{}, len(got))
	for _, route := range got {
		if _, duplicate := byID[route.ID]; duplicate {
			t.Fatalf("duplicate route id %s", route.ID)
		}
		key := route.Method + " " + route.RuntimePattern
		if _, duplicate := keys[key]; duplicate {
			t.Fatalf("duplicate runtime route %s", key)
		}
		byID[route.ID], keys[key] = route, struct{}{}
	}
	for _, expected := range want {
		actual, found := byID[expected.ID]
		runtimePath := strings.ReplaceAll(expected.Path, ":id", "{id}")
		if !found || actual.Method != expected.Method ||
			actual.LegacyPath != expected.Path || actual.RuntimePattern != runtimePath {
			t.Fatalf("route %s mismatch: %#v", expected.ID, actual)
		}
	}
}

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

func readCharacterizedRoutes(t *testing.T) []characterizedRoute {
	t.Helper()
	directory := filepath.Join(repositoryRoot(t), "contracts", "characterization",
		"domains", "billing-v1", "routes")
	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil || len(files) != 6 {
		t.Fatalf("characterization route files: %d %v", len(files), err)
	}
	result := make([]characterizedRoute, 0, 26)
	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var document struct {
			Routes []characterizedRoute `json:"routes"`
		}
		if err := json.Unmarshal(content, &document); err != nil {
			t.Fatal(err)
		}
		result = append(result, document.Routes...)
	}
	return result
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
