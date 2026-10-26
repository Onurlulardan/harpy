package catalogowner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	nativequery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/nativequery"
)

type analyticsFixtureGroup struct {
	Payloads []struct {
		ComponentID string          `json:"component_id"`
		ContentSHA  string          `json:"content_sha256"`
		Contract    string          `json:"contract"`
		Document    json.RawMessage `json:"document"`
	} `json:"payloads"`
}

func TestFirstPartyRuntimePlansHaveExplicitOperationalOutcome(t *testing.T) {
	root := repositoryRoot(t)
	files, err := filepath.Glob(filepath.Join(
		root, "tools", "internal", "firstpartycatalog", "testdata",
		"*", "artifacts", "owner-analytics.json",
	))
	if err != nil || len(files) != 8 {
		t.Fatalf("analytics fixtures: files=%d err=%v", len(files), err)
	}
	counts := map[string]int{}
	blockers := map[string]int{}
	nativeQueries := 0
	for _, file := range files {
		document, readErr := os.ReadFile(file)
		if readErr != nil {
			t.Fatal(readErr)
		}
		var group analyticsFixtureGroup
		if json.Unmarshal(document, &group) != nil {
			t.Fatalf("decode %s", file)
		}
		for _, payload := range group.Payloads {
			component := contract.Component{
				ID: payload.ComponentID, Owner: owner, Generation: 1,
				Contract: payload.Contract, ContentSHA256: payload.ContentSHA,
			}
			planJSON, digest, buildErr := buildRuntimePlan(
				payload.Document, component,
			)
			if buildErr != nil || digest == [32]byte{} {
				t.Fatalf("%s/%s: %v", file, payload.ComponentID, buildErr)
			}
			var plan runtimePlan
			if json.Unmarshal(planJSON, &plan) != nil {
				t.Fatalf("decode runtime plan %s/%s", file, payload.ComponentID)
			}
			counts[plan.Kind+":"+plan.Status]++
			if plan.BlockerCode != "" {
				blockers[plan.BlockerCode]++
			}
			if _, native, parseErr := nativequery.Parse(plan.Query); parseErr != nil {
				t.Fatalf("parse native query %s/%s: %v",
					file, payload.ComponentID, parseErr)
			} else if native {
				nativeQueries++
			}
		}
	}
	if got := counts["endpoint:candidate"]; got != 77 {
		t.Fatalf("candidate endpoints=%d, want 77; counts=%v", got, counts)
	}
	if got := counts["endpoint:pending"]; got != 0 {
		t.Fatalf("pending endpoints=%d, want 0; counts=%v", got, counts)
	}
	if got := counts["dashboard:candidate"]; got != 16 {
		t.Fatalf("candidate dashboards=%d, want 16; counts=%v", got, counts)
	}
	if nativeQueries != 6 {
		t.Fatalf("native join queries=%d, want 6", nativeQueries)
	}
	if len(blockers) != 0 {
		t.Fatalf("unexpected blockers: %v", blockers)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	current := filepath.Dir(source)
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("repository root unavailable")
		}
		current = parent
	}
}
