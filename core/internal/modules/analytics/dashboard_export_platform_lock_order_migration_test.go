package analytics

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestDashboardExportPlatformWrappersUseOnePlatformFirstLockOrder(t *testing.T) {
	wrappers := []struct {
		file, function, signature, platform, tenantLock string
	}{
		{"20260802010015_create_analytics_dashboard_export_platform_claim_v1",
			"analytics_dashboard_export_platform_claim_v1",
			"name,uuid,uuid,uuid,bigint,uuid",
			"platform_async_analytics_dashboard_export_claim_v1", "FOR UPDATE"},
		{"20260802010016_create_analytics_dashboard_export_platform_lease_v1",
			"analytics_dashboard_export_platform_checkpoint_v1",
			"name,uuid,uuid,uuid,uuid,bigint,uuid,timestamptz",
			"platform_async_analytics_dashboard_export_checkpoint_v1", "FOR UPDATE"},
		{"20260802010016_create_analytics_dashboard_export_platform_lease_v1",
			"analytics_dashboard_export_platform_effect_start_v1",
			"name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,text,text,bytea,bytea," +
				"bytea,bytea,bytea,bigint,bigint,bigint,text,text,timestamptz",
			"platform_async_analytics_dashboard_export_effect_start_v1", "FOR UPDATE"},
		{"20260802010017_create_analytics_dashboard_export_platform_succeed_v1",
			"analytics_dashboard_export_platform_succeed_v1",
			"name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,bytea,text,text," +
				"timestamptz,timestamptz",
			"platform_async_analytics_dashboard_export_succeed_v1", "FOR UPDATE"},
		{"20260802010018_create_analytics_dashboard_export_platform_finish_v1",
			"analytics_dashboard_export_platform_finish_v1",
			"name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,text,text,bytea," +
				"boolean,text,timestamptz",
			"platform_async_analytics_dashboard_export_finish_v1", "FOR UPDATE"},
		{"20260802010019_create_analytics_dashboard_export_platform_unknown_v1",
			"analytics_dashboard_export_platform_unknown_v1",
			"name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz",
			"platform_async_analytics_dashboard_export_unknown_v1", "FOR UPDATE"},
		{"20260802010019_create_analytics_dashboard_export_platform_unknown_v1",
			"analytics_dashboard_export_platform_spool_active_v1",
			"name,uuid,uuid,bigint,uuid",
			"platform_async_analytics_dashboard_export_spool_state_v1", "FOR SHARE"},
	}
	files := map[string][]byte{}
	for _, wrapper := range wrappers {
		source, exists := files[wrapper.file]
		if !exists {
			source = readDashboardExportPlatformMigration(t, "fresh", wrapper.file)
			adopted := readDashboardExportPlatformMigration(t, "v1_adopted", wrapper.file)
			if !bytes.Equal(source, adopted) {
				t.Fatalf("dashboard export platform migration pair differs: %s", wrapper.file)
			}
			lineCount := bytes.Count(source, []byte{'\n'})
			if len(source) > 0 && source[len(source)-1] != '\n' {
				lineCount++
			}
			if lineCount > 200 {
				t.Fatalf("dashboard export platform migration exceeds 200 lines: %s", wrapper.file)
			}
			if _, err := postgresv1.InspectMigrationSource(source); err != nil {
				t.Fatalf("dashboard export platform migration %s is invalid: %v", wrapper.file, err)
			}
			if bytes.Contains(source, []byte("<>")) {
				t.Fatalf("dashboard export platform migration uses nullable raw inequality: %s", wrapper.file)
			}
			files[wrapper.file] = source
		}
		body := dashboardExportPlatformFunctionBody(t, string(source), wrapper.function)
		platformAt := strings.Index(body, wrapper.platform)
		lockAt := strings.Index(body, wrapper.tenantLock)
		if platformAt < 0 || lockAt < 0 || platformAt >= lockAt {
			t.Fatalf("%s lock order is not Platform first: platform=%d tenant=%d",
				wrapper.function, platformAt, lockAt)
		}
		if strings.Count(body, wrapper.platform) != 1 {
			t.Fatalf("%s must invoke its Platform transition exactly once", wrapper.function)
		}
		normalized := strings.Join(strings.Fields(string(source)), "")
		grant := "GRANTEXECUTEONFUNCTION" + wrapper.function + "(" +
			wrapper.signature + ")TOharpy_runtime;"
		revoke := "REVOKEALLONFUNCTION" + wrapper.function + "(" +
			wrapper.signature + ")FROMPUBLIC;"
		if !strings.Contains(normalized, grant) || !strings.Contains(normalized, revoke) {
			t.Fatalf("%s does not expose only its owner-bound runtime signature", wrapper.function)
		}
	}
	unknown := strings.Join(strings.Fields(string(files["20260802010019_create_analytics_dashboard_export_platform_unknown_v1"])), "")
	for _, legacy := range []string{
		"analytics_dashboard_export_claim_v1(name,uuid,uuid,uuid,timestamptz)",
		"analytics_dashboard_export_effect_start_v1(name,uuid,uuid,bigint,text,text," +
			"bytea,bytea,bytea,bytea,bytea,bigint,bigint,bigint,text,text,timestamptz)",
		"analytics_dashboard_export_complete_v1(name,uuid,uuid,bigint,text,text," +
			"timestamptz,timestamptz)",
		"analytics_dashboard_export_restart_missing_v1(name,uuid,uuid,bigint,timestamptz)",
		"analytics_dashboard_export_outcome_unknown_v1(name,uuid,uuid,bigint,timestamptz)",
		"analytics_dashboard_export_fail_render_v1(name,uuid,uuid,bigint,text," +
			"boolean,timestamptz)",
	} {
		if !strings.Contains(unknown, "REVOKEEXECUTEONFUNCTION"+legacy+
			"FROMharpy_runtime;") {
			t.Fatalf("legacy dashboard export ABI remains executable: %s", legacy)
		}
	}
}

func readDashboardExportPlatformMigration(t *testing.T, origin, logical string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("migrations", "tenant", origin,
		logical+"."+origin+".sql"))
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func dashboardExportPlatformFunctionBody(t *testing.T, source, function string) string {
	t.Helper()
	start := strings.Index(source, "CREATE FUNCTION "+function+"(")
	if start < 0 {
		t.Fatalf("dashboard export Platform wrapper is missing: %s", function)
	}
	end := strings.Index(source[start:], "$function$;")
	if end < 0 {
		t.Fatalf("dashboard export Platform wrapper is unterminated: %s", function)
	}
	return source[start : start+end]
}
