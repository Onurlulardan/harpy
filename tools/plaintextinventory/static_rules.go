package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type staticKind uint8

const (
	staticSkip staticKind = iota
	staticDeployment
	staticSource
	staticForgeGuardSource
)

var (
	legacyBootstrap = regexp.MustCompile(`(?i)\b(SUPER_ADMIN_(PASSWORD|EMAIL)|` +
		`DEFAULT_TENANT_(ID|NAME|SLUG)|DEFAULT_ADMIN_PASSWORD|HARPY_LICENSE_KEY)\b`)
	liveEnvLookup = regexp.MustCompile(`(?i)(Getenv|LookupEnv|process\.env|import\.meta\.env|` +
		`environment\[)[^\n]{0,100}(SUPER_ADMIN|DEFAULT_TENANT|HARPY_LICENSE_KEY)`)
	legacyAdminBoolean = regexp.MustCompile(`(?i)\b(is_super_admin|isSuperAdmin)\b`)
	browserSecretStore = regexp.MustCompile(`(?i)(localStorage|sessionStorage)[^\n]{0,160}` +
		`(password|initial.admin|credential|super.admin)`)
)

func staticFileKind(relative string, entry os.DirEntry) staticKind {
	if entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
		return staticSkip
	}
	lower := strings.ToLower(relative)
	base := filepath.Base(lower)
	if strings.Contains(base, "_test.") || strings.HasSuffix(base, ".md") ||
		strings.HasSuffix(base, ".lock") || base == "routetree.gen.ts" {
		return staticSkip
	}
	if deploymentFile(lower, base) {
		return staticDeployment
	}
	if sourceFile(lower) {
		if strings.HasPrefix(lower, "forgeguard/") {
			return staticForgeGuardSource
		}
		return staticSource
	}
	return staticSkip
}

func deploymentFile(path, base string) bool {
	extension := filepath.Ext(base)
	configured := map[string]bool{
		".yml": true, ".yaml": true, ".env": true, ".json": true,
		".toml": true, ".ini": true, ".conf": true, ".sh": true,
		".ps1": true, ".bat": true, ".cmd": true, ".tmpl": true,
	}
	isConfigured := configured[extension] || strings.Contains(base, ".env")
	return isConfigured &&
		(strings.Contains(path, "/ops/") || strings.Contains(path, "/deploy/") ||
			strings.Contains(path, "docker-compose") || strings.Contains(base, ".env")) ||
		strings.Contains(path, "/composegen/") && extension == ".go"
}

func sourceFile(path string) bool {
	extension := filepath.Ext(path)
	if extension != ".go" && extension != ".ts" && extension != ".tsx" &&
		extension != ".js" && extension != ".jsx" {
		return false
	}
	return strings.HasPrefix(path, "backend-v2/core/") ||
		strings.HasPrefix(path, "frontend-v3/src/") ||
		strings.HasPrefix(path, "forgeguard/backend/internal/") ||
		strings.HasPrefix(path, "forgeguard/frontend/src/")
}

func staticLineViolation(line string, kind staticKind) string {
	if kind == staticDeployment && legacyBootstrap.MatchString(line) {
		return "legacy bootstrap plaintext configuration"
	}
	if (kind == staticSource || kind == staticForgeGuardSource) && liveEnvLookup.MatchString(line) {
		return "legacy bootstrap environment lookup"
	}
	if kind == staticSource && legacyAdminBoolean.MatchString(line) {
		return "super-admin boolean authority"
	}
	if (kind == staticSource || kind == staticForgeGuardSource) && browserSecretStore.MatchString(line) {
		return "browser installation-secret persistence"
	}
	return ""
}
