package asyncadoptionv1

import (
	"encoding/json"
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

func TestV1DifferentialPinsActualSourceSnapshot(t *testing.T) {
	document := readV1Fixture(t, "v1-source-snapshot.json")
	var sources []struct{ Path, SHA256 string }
	if err := json.Unmarshal(document, &sources); err != nil {
		t.Fatal(err)
	}
	paths := map[string]struct{}{}
	pattern := regexp.MustCompile(`^[0-9a-f]{64}$`)
	for _, source := range sources {
		if !fs.ValidPath(source.Path) || !strings.HasPrefix(source.Path, "backend/") ||
			!pattern.MatchString(source.SHA256) {
			t.Fatalf("invalid V1 source evidence: %#v", source)
		}
		if _, duplicate := paths[source.Path]; duplicate {
			t.Fatalf("duplicate V1 source evidence: %s", source.Path)
		}
		paths[source.Path] = struct{}{}
	}
	if len(sources) != 3 || len(paths) != len(sources) {
		t.Fatalf("V1 source snapshot count = %d/%d", len(sources), len(paths))
	}
}
