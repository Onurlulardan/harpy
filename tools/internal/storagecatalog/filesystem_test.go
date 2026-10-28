package storagecatalog

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
		writeRoot(t, root, catalogVersion, []string{ownerPath("iam")})
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "read storage owner fragment") {
			t.Fatalf("missing fragment error = %v", err)
		}
	})
	t.Run("orphan", func(t *testing.T) {
		root := t.TempDir()
		writeRoot(t, root, catalogVersion, []string{ownerPath("iam")})
		writeOwner(t, root, ownerDocumentVersion, validOwner("iam"))
		writeOwner(t, root, ownerDocumentVersion, validOwner("zeta"))
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "orphan") {
			t.Fatalf("orphan fragment error = %v", err)
		}
	})
}

func TestLoadRejectsNonRegularAndSymlinkFragments(t *testing.T) {
	t.Run("non_regular", func(t *testing.T) {
		root := t.TempDir()
		include := ownerPath("iam")
		writeRoot(t, root, catalogVersion, []string{include})
		if err := os.MkdirAll(
			filepath.Join(root, filepath.FromSlash(include)), 0o700,
		); err != nil {
			t.Fatal(err)
		}
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "regular file") {
			t.Fatalf("non-regular fragment error = %v", err)
		}
	})
	t.Run("file_symlink", func(t *testing.T) {
		root := t.TempDir()
		include := ownerPath("iam")
		writeRoot(t, root, catalogVersion, []string{include})
		target := filepath.Join(root, "outside.json")
		content, err := json.Marshal(ownerDocument{
			Version: ownerDocumentVersion, Owner: validOwner("iam"),
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

func TestLoadRejectsSymlinkOwnerDirectory(t *testing.T) {
	root := t.TempDir()
	writeRoot(t, root, catalogVersion, []string{ownerPath("iam")})
	target := t.TempDir()
	writeJSONFile(t, target, "iam.json", ownerDocument{
		Version: ownerDocumentVersion, Owner: validOwner("iam"),
	})
	link := filepath.Join(root, filepath.FromSlash(ownerDirectoryPath))
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

func TestLoadRejectsRepositoryParentSymlink(t *testing.T) {
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

func TestOpenInspectedFileRejectsPathReplacement(t *testing.T) {
	root := t.TempDir()
	writeRoot(t, root, catalogVersion, []string{})
	repository, err := openRepositoryRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	expected, err := inspectRepositoryPath(repository, rootDocumentPath)
	if err != nil {
		t.Fatal(err)
	}
	filename := filepath.Join(root, filepath.FromSlash(rootDocumentPath))
	if err := os.Rename(filename, filename+".old"); err != nil {
		t.Fatal(err)
	}
	writeRawFile(t, root, rootDocumentPath, []byte(`{"version":2,"includes":[]}`))
	file, err := openInspectedFile(repository, rootDocumentPath, expected)
	if file != nil {
		file.Close()
	}
	if err == nil || !strings.Contains(err.Error(), "changed while opening") {
		t.Fatalf("replacement error = %v", err)
	}
}
