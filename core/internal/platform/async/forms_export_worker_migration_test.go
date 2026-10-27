package async

import (
	"bytes"
	"path"
	"strconv"
	"strings"
	"testing"
)

func TestFormsExportWorkerPlatformMigrationContract(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	pairs := map[string]map[string][]byte{}
	for _, document := range documents {
		name := path.Base(document.Path())
		version, parseErr := strconv.ParseInt(name[:14], 10, 64)
		if parseErr != nil || version < 20260718020035 || version > 20260718020049 {
			continue
		}
		for _, origin := range []string{"fresh", "v1_adopted"} {
			suffix := "." + origin + ".sql"
			if !strings.HasSuffix(name, suffix) {
				continue
			}
			logical := strings.TrimSuffix(name, suffix)
			if pairs[logical] == nil {
				pairs[logical] = map[string][]byte{}
			}
			pairs[logical][origin] = document.Content()
		}
	}
	if len(pairs) != 11 {
		t.Fatalf("Forms export worker platform pairs=%d, want 11", len(pairs))
	}
	for logical, pair := range pairs {
		if len(pair["fresh"]) == 0 || !bytes.Equal(pair["fresh"], pair["v1_adopted"]) {
			t.Fatalf("Forms export worker platform pair differs: %s", logical)
		}
	}
	claim := string(pairs["20260718020036_create_forms_export_claim_v1"]["fresh"])
	for _, token := range []string{"2147483648", "quota_wait", "job_attempts",
		"attempts=p_attempt_generation"} {
		if !strings.Contains(claim, token) {
			t.Fatalf("Forms export platform claim migration misses %q", token)
		}
	}
	unknown := string(pairs["20260718020044_allow_current_external_unknown_v1"]["fresh"])
	for _, token := range []string{"v_authority_current",
		"platform_async_validate_stale_external_attempt_unknown_v1",
		"platform_async_validate_stale_external_job_unknown_v1"} {
		if !strings.Contains(unknown, token) {
			t.Fatalf("Forms export unknown guard migration misses %q", token)
		}
	}
}
