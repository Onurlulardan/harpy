package asyncadoptionv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
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
	workspaceRoot, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range sources {
		if !fs.ValidPath(source.Path) || !strings.HasPrefix(source.Path, "backend/") ||
			!pattern.MatchString(source.SHA256) {
			t.Fatalf("invalid V1 source evidence: %#v", source)
		}
		if _, duplicate := paths[source.Path]; duplicate {
			t.Fatalf("duplicate V1 source evidence: %s", source.Path)
		}
		paths[source.Path] = struct{}{}
		content, err := os.ReadFile(filepath.Join(
			workspaceRoot, filepath.FromSlash(source.Path),
		))
		if err != nil {
			t.Fatalf("read pinned V1 source %s: %v", source.Path, err)
		}
		digest := sha256.Sum256(content)
		if actual := hex.EncodeToString(digest[:]); actual != source.SHA256 {
			t.Fatalf("V1 source %s SHA-256 = %s, want %s",
				source.Path, actual, source.SHA256)
		}
	}
	if len(sources) != 8 || len(paths) != len(sources) {
		t.Fatalf("V1 source snapshot count = %d/%d", len(sources), len(paths))
	}
}
