package goose

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

// Snapshot names use ".fresh.sql". Stripping the routine word "refresh" keeps that
// marker without rejecting billing_license_refresh_* migrations.
var legacyRoleStatement = regexp.MustCompile(`(?i)CREATE\s+ROLE\s+(?:IF\s+NOT\s+EXISTS\s+)?("?[A-Za-z_][A-Za-z0-9_]*"?)`)

func TestSourceOmitsLegacyCorpus(t *testing.T) {
	for _, root := range []struct {
		files fs.FS
		dir   string
	}{{platformRoot, "sql/platform"}, {tenantRoot, "sql/tenant"}} {
		source, err := fs.Sub(root.files, root.dir)
		if err != nil {
			t.Fatal(err)
		}
		err = fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() {
				return walkErr
			}
			if legacyMigrationName(path) {
				t.Fatalf("legacy migration name %s", path)
			}
			body, err := fs.ReadFile(source, path)
			if err != nil {
				return err
			}
			// PRESERVE_NON_MIGRATION: Goose SQL must not create the native history table itself.
			if legacyMigrationBody(string(body)) {
				t.Fatalf("legacy body in %s", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestLegacyCorpusMarkersStayClosed(t *testing.T) {
	for _, path := range []string{
		"20261028000000_platform.sql",
		"20261028000001_tenant_baseline.fresh.sql",
		"20261102000000_catalog.sql",
		"20261103000003_register.sql",
		"20261103000004_claim.sql",
		"20261103000005_heartbeat.sql",
		"20261103000006_complete.sql",
		"fresh.sql",
		"baseline.fresh.sql",
	} {
		if !legacyMigrationName(path) {
			t.Fatalf("legacy migration name accepted: %s", path)
		}
	}
	if legacyMigrationName("00041_billing_license_refresh_claim_v1.sql") {
		t.Fatal("refresh routine filename rejected")
	}

	for _, body := range []string{
		"CREATE FUNCTION platform.billing_license_clock_read_v1()",
		"CREATE OR REPLACE FUNCTION platform.example_v1()",
		"CREATE PROCEDURE platform.example_v1()",
		"CREATE TRIGGER example_touch BEFORE UPDATE ON platform.example",
		"CREATE VIEW platform.example_v1 AS SELECT 1",
		"CREATE ROLE harpy_api NOLOGIN",
		"CREATE ROLE harpy_backup_reader NOLOGIN",
	} {
		if legacyMigrationBody(body) {
			t.Fatalf("goose object rejected: %s", body)
		}
	}
	for _, body := range []string{
		"CREATE TABLE goose_db_version (version_id bigint)",
		"-- copied from 20261028000000_platform.fresh.sql",
		"-- removed 20261103000003 through 20261103000006",
		"CREATE ROLE harpy_v2_runtime LOGIN",
		"CREATE ROLE harpy_runtime NOLOGIN",
		"CREATE ROLE IF NOT EXISTS harpy_cutover_decoder NOLOGIN",
		"CREATE ROLE harpy_cutover_authority_owner NOLOGIN",
		"CREATE ROLE harpy_adopter LOGIN",
		"CREATE ROLE harpy_backup LOGIN",
		`CREATE ROLE "harpy_runtime" NOLOGIN`,
	} {
		if !legacyMigrationBody(body) {
			t.Fatalf("legacy body accepted: %s", body)
		}
	}
}

func legacyMigrationName(path string) bool {
	return legacyCorpusMarker(path)
}

func legacyMigrationBody(text string) bool {
	if strings.Contains(text, "goose_db_version") || legacyCorpusMarker(text) || forbiddenRuntimeRole(text) {
		return true
	}
	return false
}

func legacyCorpusMarker(value string) bool {
	if strings.Contains(value, "20261028") || strings.Contains(value, "20261102") || strings.Contains(value, "20261103") {
		return true
	}
	return strings.Contains(strings.ReplaceAll(value, "refresh", ""), "fresh")
}

func forbiddenRuntimeRole(text string) bool {
	for _, match := range legacyRoleStatement.FindAllStringSubmatch(text, -1) {
		name := strings.ToLower(strings.Trim(match[1], `"`))
		switch name {
		case "harpy_v2_runtime", "harpy_runtime", "harpy_adopter", "harpy_backup":
			return true
		}
		if strings.HasPrefix(name, "harpy_cutover_") {
			return true
		}
	}
	return false
}
