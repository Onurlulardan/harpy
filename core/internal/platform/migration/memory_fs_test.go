package migration

import (
	"io/fs"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestMemoryFSIsReadOnlyDefensiveAndDeterministic(t *testing.T) {
	first := []byte("first")
	source := newMemoryFS(map[string][]byte{
		"z/file.sql":         []byte("last"),
		"a/one.sql":          first,
		"a/two.sql":          []byte("second"),
		"a/nested/three.sql": []byte("third"),
	})
	first[0] = 'X'
	content, err := fs.ReadFile(source, "a/one.sql")
	if err != nil || string(content) != "first" {
		t.Fatalf("memory filesystem did not defend source bytes: %q %v", content, err)
	}
	root, err := fs.ReadDir(source, ".")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range root {
		names = append(names, entry.Name())
	}
	if !reflect.DeepEqual(names, []string{"a", "z"}) {
		t.Fatalf("root entries are not deterministic: %v", names)
	}
	entries, err := fs.ReadDir(source, "a")
	if err != nil || entries[0].Name() != "nested" || entries[1].Name() != "one.sql" ||
		entries[2].Name() != "two.sql" {
		t.Fatalf("directory entries are not sorted: %v %v", entries, err)
	}
	if _, err := source.Open("missing.sql"); err == nil {
		t.Fatal("missing file did not return an error")
	}
}

func TestNewCatalogRejectsUnverifiedArtifacts(t *testing.T) {
	if _, err := NewCatalog(nil, attestation.VerifiedPostgresArtifactSet{}, nil); err == nil {
		t.Fatal("unverified PostgreSQL artifacts were accepted")
	}
}
