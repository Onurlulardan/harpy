package ddl

import (
	"strings"
	"testing"
)

func TestRenderBackfillV2UsesBoundedRecordPageEvidence(t *testing.T) {
	query := renderBackfillV2PageSQL(`"tenant_a"."frm_orders"`,
		`"fld_source"`, `"fld_target"`, "selected.source_value")
	for _, required := range []string{
		"ORDER BY id LIMIT 1001 FOR UPDATE",
		"row_bytes>1048576 OR position>1000 OR cumulative_bytes>8388608",
		"measured.position<COALESCE(boundary.position,1001)",
		"pg_catalog.octet_length((SELECT COALESCE(pg_catalog.string_agg(",
		"pg_catalog.encode(pg_catalog.sha256(",
		"COALESCE((SELECT row_bytes FROM measured,boundary",
	} {
		if !strings.Contains(query, required) {
			t.Fatalf("V2 backfill query lacks %q", required)
		}
	}
}
