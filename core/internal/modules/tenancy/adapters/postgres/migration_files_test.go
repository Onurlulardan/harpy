package postgres

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const (
	firstTenancyMigration   = int64(20260713030000)
	lastTenancyMigration    = int64(20260713039999)
	runtimeTenancyMigration = int64(20260718050600)
	phase14TenancyFirst     = int64(20260722010000)
	phase14TenancyLast      = int64(20260722142010)
	v1AdoptionFirst         = int64(20260723112000)
	v1AdoptionLast          = int64(20260723115000)
	phase14ClosureFirst     = int64(20260725120000)
	phase14ClosureLast      = int64(20260728103010)
	generation70First       = int64(20260729100002)
	generation70Last        = int64(20260729100033)
	generation71First       = int64(20260729100037)
	generation71Last        = int64(20260729100037)
	tenantSettingsRuntime   = int64(20260801020000)
	customerLifecycleFirst  = int64(20260802232300)
	customerLifecycleLast   = int64(20260802232305)
)

func platformMigrationNames(t *testing.T, origin string) []string {
	t.Helper()
	entries, err := os.ReadDir(platformMigrationDirectory(origin))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("unexpected migration subdirectory: %s", entry.Name())
		}
		names = append(names, entry.Name())
	}
	return names
}

func assertTenancyMigrationVersion(t *testing.T, name string) {
	t.Helper()
	if len(name) < 16 || name[14] != '_' {
		t.Fatalf("invalid migration basename: %q", name)
	}
	version, err := strconv.ParseInt(name[:14], 10, 64)
	legacyRange := version >= firstTenancyMigration && version <= lastTenancyMigration
	phase14Range := version >= phase14TenancyFirst && version <= phase14TenancyLast
	v1AdoptionRange := version >= v1AdoptionFirst && version <= v1AdoptionLast
	phase14ClosureRange := version >= phase14ClosureFirst && version <= phase14ClosureLast
	generation70Range := version >= generation70First && version <= generation70Last
	generation71Range := version >= generation71First && version <= generation71Last
	customerLifecycleRange := version >= customerLifecycleFirst && version <= customerLifecycleLast
	if err != nil || !legacyRange && version != runtimeTenancyMigration &&
		!phase14Range && !v1AdoptionRange && !phase14ClosureRange &&
		!generation70Range && !generation71Range && !customerLifecycleRange &&
		version != tenantSettingsRuntime {
		t.Fatalf("migration version outside tenancy range: %q", name)
	}
}

func assertMigrationSourcePolicy(t *testing.T, origin, name string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(platformMigrationDirectory(origin), name))
	if err != nil {
		t.Fatal(err)
	}
	upper := strings.ToUpper(string(content))
	for _, forbidden := range []string{"-- +GOOSE DOWN", "DO $$", "\\I "} {
		if strings.Contains(upper, forbidden) {
			t.Fatalf("%s contains forbidden token %q", name, forbidden)
		}
	}
}

func platformMigrationDirectory(origin string) string {
	return filepath.Join("..", "..", "migrations", "platform", origin)
}
