package migrationstreamcatalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRejectsMissingAndOrphanFragments(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		root := t.TempDir()
		writeRoot(t, root, catalogVersion, []string{streamPath("iam")})
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "read migration stream fragment") {
			t.Fatalf("missing fragment error = %v", err)
		}
	})
	t.Run("orphan", func(t *testing.T) {
		root := t.TempDir()
		writeRoot(t, root, catalogVersion, []string{streamPath("iam")})
		writeOwner(t, root, ownerDocumentVersion, "iam", []Stream{validStream("iam", 1)})
		writeOwner(t, root, ownerDocumentVersion, "zeta", []Stream{validStream("zeta", 2)})
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "orphan") {
			t.Fatalf("orphan fragment error = %v", err)
		}
	})
}

func TestLoadRejectsNonRegularAndSymlinkFragments(t *testing.T) {
	t.Run("non_regular", func(t *testing.T) {
		root := t.TempDir()
		include := streamPath("iam")
		writeRoot(t, root, catalogVersion, []string{include})
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(include)), 0o700); err != nil {
			t.Fatal(err)
		}
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "regular file") {
			t.Fatalf("non-regular fragment error = %v", err)
		}
	})
	t.Run("file_symlink", func(t *testing.T) {
		root := t.TempDir()
		include := streamPath("iam")
		writeRoot(t, root, catalogVersion, []string{include})
		target := filepath.Join(root, "outside.json")
		content, err := json.Marshal(ownerDocument{
			Version: ownerDocumentVersion, Owner: "iam",
			Streams: []Stream{validStream("iam", 1)},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, content, 0o600); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(root, filepath.FromSlash(include))
		if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, link); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		_, err = Load(root)
		if err == nil || !strings.Contains(err.Error(), "symbolic link") {
			t.Fatalf("symlink fragment error = %v", err)
		}
	})
}

func TestLoadRejectsSymlinkFragmentDirectory(t *testing.T) {
	root := t.TempDir()
	writeRoot(t, root, catalogVersion, []string{streamPath("iam")})
	target := t.TempDir()
	writeJSONFile(t, target, "iam.json", ownerDocument{
		Version: ownerDocumentVersion, Owner: "iam",
		Streams: []Stream{validStream("iam", 1)},
	})
	link := filepath.Join(root, filepath.FromSlash(streamDirectoryPath))
	if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("directory symlink unavailable: %v", err)
	}
	_, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("symlink directory error = %v", err)
	}
}

func TestLoadRejectsRepositoryRootOrParentSymlink(t *testing.T) {
	targetParent := t.TempDir()
	realRoot := filepath.Join(targetParent, "repository")
	if err := os.Mkdir(realRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	writeRoot(t, realRoot, catalogVersion, []string{})
	linkParent := filepath.Join(t.TempDir(), "linked_parent")
	if err := os.Symlink(targetParent, linkParent); err != nil {
		t.Skipf("parent symlink unavailable: %v", err)
	}
	_, err := Load(filepath.Join(linkParent, "repository"))
	if err == nil || !strings.Contains(err.Error(), "contains a symbolic link") {
		t.Fatalf("repository parent symlink error = %v", err)
	}
}
