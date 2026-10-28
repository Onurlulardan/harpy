package migrationstreamcatalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func validStream(owner string, version int64) Stream {
	return Stream{
		Owner: owner, Directory: "core/internal/" + owner + "/migrations",
		Scope: "platform", VersionFrom: version, VersionTo: version,
		Phase: "foundation", TargetDatabaseGeneration: 1,
		NextMinimumRuntimeGeneration: 1,
	}
}

func streamPath(owner string) string {
	return streamDirectoryPath + "/" + owner + ".json"
}

func writeRoot(t *testing.T, root string, version int, includes []string) {
	t.Helper()
	writeJSONFile(t, root, rootDocumentPath, rootDocument{
		Version: version, Includes: includes,
	})
}

func writeOwner(t *testing.T, root string, version int, owner string, streams []Stream) {
	t.Helper()
	writeJSONFile(t, root, streamPath(owner), ownerDocument{
		Version: version, Owner: owner, Streams: streams,
	})
}

func writeJSONFile(t *testing.T, root, relative string, value any) {
	t.Helper()
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeRawFile(t, root, relative, content)
}

func writeRawFile(t *testing.T, root, relative string, content []byte) {
	t.Helper()
	filename := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(filename), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, content, 0o600); err != nil {
		t.Fatal(err)
	}
}
