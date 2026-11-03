package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

func pinnedAllowlist() []string {
	return []string{
		"go.mod",
		".github/workflows/ci.yml",
		"contracts/runtime/v1/identity.go",
		"contracts/runtime/v1/identity_test.go",
		"contracts/runtime/v1/version.schema.json",
		"contracts/runtime/v1/version_test.go",
		"ops/local/Dockerfile",
		"ops/production/Dockerfile",
		"ops/characterization/preflight.sh",
		"ops/characterization/README.md",
		"README.md",
		"docs/architecture/01-genel-bakis.md",
		"docs/architecture/10-test-ve-muhendislik-standartlari.md",
		"docs/architecture/11-gelistirici-rehberi.md",
	}
}

func loadPinnedFiles(root string, current goRelease) (map[string]string, error) {
	found := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			if entry.Name() == ".git" || filepath.ToSlash(relative) == "ops/local/state" {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Size() > 2<<20 {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if strings.HasPrefix(relative, "ops/local/state/") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !utf8.Valid(content) || !mentionsRelease(string(content), current) {
			return nil
		}
		found[relative] = string(content)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan Go pins: %w", err)
	}
	return found, classifyPins(found, pinnedAllowlist())
}

func classifyPins(found map[string]string, allowed []string) error {
	allow := make(map[string]bool, len(allowed))
	for _, path := range allowed {
		allow[path] = true
		if _, ok := found[path]; !ok {
			return fmt.Errorf("pinned file %s does not contain the current Go release", path)
		}
	}
	for path := range found {
		if !allow[path] {
			return fmt.Errorf("unexpected Go pin in %s", path)
		}
	}
	return nil
}

func mentionsRelease(content string, release goRelease) bool {
	return containsVersion(content, release.String()) ||
		containsVersion(content, release.Language()) ||
		strings.Contains(content, release.Toolchain())
}

func containsVersion(content, version string) bool {
	pattern := `(?:^|[^0-9])` + regexp.QuoteMeta(version) + `(?:[^0-9]|$)`
	return regexp.MustCompile(pattern).MatchString(content)
}
