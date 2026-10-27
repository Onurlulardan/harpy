package async

import (
	"bytes"
	"path"
	"strconv"
	"strings"
	"testing"
)

func TestFormsImportPlatformMigrationContract(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	pairs := map[string]map[string][]byte{}
	for _, document := range documents {
		name := path.Base(document.Path())
		version, parseErr := strconv.ParseInt(name[:14], 10, 64)
		if parseErr != nil || version < 20260718020050 || version > 20260718020059 {
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
	if len(pairs) != 7 {
		t.Fatalf("Forms import platform pairs=%d, want 7", len(pairs))
	}
	for logical, pair := range pairs {
		if len(pair["fresh"]) == 0 || !bytes.Equal(pair["fresh"], pair["v1_adopted"]) {
			t.Fatalf("Forms import platform pair differs: %s", logical)
		}
	}
	submit := string(pairs["20260718020051_create_forms_import_submit_v1"]["fresh"])
	for _, token := range []string{
		"forms.record-import", "'database_transactional'", "0,8,0",
		"forms.records.import", "forms-record-import/",
	} {
		if !strings.Contains(submit, token) {
			t.Fatalf("Forms import submit migration misses %q", token)
		}
	}
}
