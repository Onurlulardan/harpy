package main

import (
	"fmt"
	"regexp"
	"strings"
)

func rewritePins(
	files map[string]string,
	current, next goRelease,
	digest string,
) (map[string]string, error) {
	if !validImageDigest(digest) {
		return nil, fmt.Errorf("image digest must be an exact sha256")
	}
	if sameRelease(current, next) {
		return nil, fmt.Errorf("requested release is already pinned")
	}
	rewritten := make(map[string]string, len(files))
	for path, content := range files {
		updated, err := rewritePin(path, content, current, next, digest)
		if err != nil {
			return nil, err
		}
		if leavesCurrentRelease(updated, current, next) {
			return nil, fmt.Errorf("current Go release remains in %s", path)
		}
		rewritten[path] = updated
	}
	return rewritten, nil
}

func rewritePin(
	path, content string,
	current, next goRelease,
	digest string,
) (string, error) {
	switch path {
	case "go.mod":
		return rewriteModule(content, current, next)
	case ".github/workflows/ci.yml":
		return replaceCount(content, path,
			`go-version: "`+current.String()+`"`,
			`go-version: "`+next.String()+`"`, 5)
	case "contracts/runtime/v1/identity.go",
		"contracts/runtime/v1/identity_test.go",
		"contracts/runtime/v1/version.schema.json",
		"contracts/runtime/v1/version_test.go":
		return replaceCount(content, path, current.Toolchain(), next.Toolchain(), 1)
	case "ops/local/Dockerfile", "ops/production/Dockerfile":
		return rewriteDockerfile(content, path, current, next, digest)
	case "ops/characterization/preflight.sh":
		return rewritePreflight(content, current, next)
	case "ops/characterization/README.md":
		return replaceCount(content, path, "Go "+current.String(), "Go "+next.String(), 1)
	case "README.md":
		return rewriteReadme(content, current, next)
	case "docs/architecture/01-genel-bakis.md":
		return replaceCount(content, path, overviewPin(current), overviewPin(next), 1)
	case "docs/architecture/10-test-ve-muhendislik-standartlari.md",
		"docs/architecture/11-gelistirici-rehberi.md":
		return replaceCount(content, path,
			"Go "+current.String()+" toolchain",
			"Go "+next.String()+" toolchain", 1)
	default:
		return "", fmt.Errorf("no rewriter for %s", path)
	}
}

func rewriteModule(content string, current, next goRelease) (string, error) {
	updated, err := replaceCount(content, "go.mod",
		"\ngo "+current.Language()+"\n", "\ngo "+next.Language()+"\n", 1)
	if err != nil {
		return "", err
	}
	return replaceCount(updated, "go.mod",
		"\ntoolchain "+current.Toolchain()+"\n",
		"\ntoolchain "+next.Toolchain()+"\n", 1)
}

func rewriteDockerfile(
	content, path string,
	current, next goRelease,
	digest string,
) (string, error) {
	pattern := regexp.MustCompile(`FROM golang:` + regexp.QuoteMeta(current.String()) +
		`-bookworm@sha256:[0-9a-f]{64} AS build`)
	if len(pattern.FindAllString(content, -1)) != 1 {
		return "", fmt.Errorf("%s does not contain one pinned bookworm Go image", path)
	}
	replacement := "FROM golang:" + next.String() + "-bookworm@" + digest + " AS build"
	return pattern.ReplaceAllString(content, replacement), nil
}

func rewritePreflight(content string, current, next goRelease) (string, error) {
	updated, err := replaceCount(content, "ops/characterization/preflight.sh",
		"*golang:"+current.String()+"*@sha256:*",
		"*golang:"+next.String()+"*@sha256:*", 1)
	if err != nil {
		return "", err
	}
	return replaceCount(updated, "ops/characterization/preflight.sh",
		"must pin Go "+current.String(), "must pin Go "+next.String(), 1)
}

func rewriteReadme(content string, current, next goRelease) (string, error) {
	updated, err := replaceCount(content, "README.md",
		"Go "+current.Minor()+"; repository toolchain "+current.String(),
		"Go "+next.Minor()+"; repository toolchain "+next.String(), 1)
	if err != nil {
		return "", err
	}
	return replaceCount(updated, "README.md",
		"Go "+current.Minor()+" toolchain (repository toolchain: "+current.String()+").",
		"Go "+next.Minor()+" toolchain (repository toolchain: "+next.String()+").", 1)
}

func overviewPin(release goRelease) string {
	return "Go `" + release.Language() + "`, toolchain `" + release.String() + "`"
}

func leavesCurrentRelease(content string, current, next goRelease) bool {
	if containsVersion(content, current.String()) || strings.Contains(content, current.Toolchain()) {
		return true
	}
	return current.Language() != next.Language() && containsVersion(content, current.Language())
}

func replaceCount(content, path, old, updated string, count int) (string, error) {
	if old == updated {
		if strings.Count(content, old) != count {
			return "", fmt.Errorf("%s does not contain %q %d times", path, old, count)
		}
		return content, nil
	}
	if strings.Count(content, old) != count {
		return "", fmt.Errorf("%s does not contain %q %d times", path, old, count)
	}
	return strings.ReplaceAll(content, old, updated), nil
}
