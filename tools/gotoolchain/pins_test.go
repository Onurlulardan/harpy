package main

import (
	"strings"
	"testing"
)

func TestRewriteRepositoryPins(t *testing.T) {
	root := moduleRootForTest(t)
	current, err := currentRelease(root)
	if err != nil {
		t.Fatal(err)
	}
	files, err := loadPinnedFiles(root, current)
	if err != nil {
		t.Fatal(err)
	}
	next := goRelease{major: 9, minor: 8, patch: 7}
	digest := "sha256:" + strings.Repeat("ab", 32)
	rewritten, err := rewritePins(files, current, next, digest)
	if err != nil {
		t.Fatal(err)
	}
	for path, content := range rewritten {
		if leavesCurrentRelease(content, current, next) {
			t.Fatalf("current release remains in %s", path)
		}
	}
	moduleFile := rewritten["go.mod"]
	if !strings.Contains(moduleFile, "\ngo "+next.Language()+"\n") ||
		!strings.Contains(moduleFile, "\ntoolchain "+next.Toolchain()+"\n") {
		t.Fatal("module toolchain was not rewritten")
	}
	image := "FROM golang:" + next.String() + "-bookworm@" + digest + " AS build"
	if !strings.Contains(rewritten["ops/local/Dockerfile"], image) ||
		!strings.Contains(rewritten["ops/production/Dockerfile"], image) {
		t.Fatal("image pin was not rewritten")
	}
	versionTest := rewritten["contracts/runtime/v1/version_test.go"]
	if strings.Contains(files["contracts/runtime/v1/version_test.go"], "go1.26.4") &&
		!strings.Contains(versionTest, "go1.26.4") {
		t.Fatal("deliberate mismatched Go version was rewritten")
	}
	if strings.Count(rewritten[".github/workflows/ci.yml"], `go-version: "`+next.String()+`"`) != 5 {
		t.Fatal("CI pins were not all rewritten")
	}
}

func TestRewriteRejectsSameRelease(t *testing.T) {
	release := goRelease{major: 1, minor: 27, patch: 3}
	_, err := rewritePins(map[string]string{}, release, release, "sha256:"+strings.Repeat("a", 64))
	if err == nil {
		t.Fatal("same release was accepted")
	}
}

func TestClassifyRejectsUnknownPin(t *testing.T) {
	found := map[string]string{"README.md": "pinned", "notes.md": "pinned"}
	if err := classifyPins(found, []string{"README.md"}); err == nil {
		t.Fatal("unknown pin was accepted")
	}
}

func moduleRootForTest(t *testing.T) string {
	t.Helper()
	root, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	return root
}
