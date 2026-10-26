package deployevidence

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOwnershipChunkReaderReadsOnlyInsideArtifactDirectory(t *testing.T) {
	base := t.TempDir()
	path := filepath.Join(base, "chunks", "000.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create chunk directory: %v", err)
	}
	if err := os.WriteFile(path, []byte("chunk"), 0o600); err != nil {
		t.Fatalf("write chunk: %v", err)
	}
	reader := &ownershipChunkReader{
		baseDirectory: base, relativePath: "chunks/000.json", maximumBytes: 5,
	}
	content, err := io.ReadAll(reader)
	if err != nil || string(content) != "chunk" {
		t.Fatalf("read chunk: content=%q err=%v", content, err)
	}
}

func TestOwnershipChunkReaderRejectsEscapingSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unprivileged Windows symlink creation is not portable")
	}
	base := t.TempDir()
	outside := writeArtifactFixture(t, []byte("outside"))
	link := filepath.Join(base, "chunk.json")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("create escaping symlink: %v", err)
	}
	reader := &ownershipChunkReader{
		baseDirectory: base, relativePath: "chunk.json", maximumBytes: 64,
	}
	_, err := io.ReadAll(reader)
	if err == nil || strings.Contains(err.Error(), outside) {
		t.Fatalf("escaping chunk result is unsafe: %v", err)
	}
}
