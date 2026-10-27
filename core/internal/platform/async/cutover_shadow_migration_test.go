package async

import (
	"bytes"
	"path"
	"strconv"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestDocumentsCutoverShadowMigrationsFreezeProtectedAuthority(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	freshDocuments := map[string][]byte{}
	for _, document := range documents {
		name := path.Base(document.Path())
		if !strings.HasSuffix(name, ".fresh.sql") {
			continue
		}
		version, parseErr := strconv.ParseInt(name[:14], 10, 64)
		if parseErr != nil || version < 20260718230700 || version > 20260718230709 {
			continue
		}
		freshDocuments[strings.TrimSuffix(name, ".fresh.sql")] = document.Content()
	}
	if len(freshDocuments) != 10 {
		t.Fatalf("fresh Documents cutover shadow migrations=%d, want 10", len(freshDocuments))
	}
	combined := make([]byte, 0, 32_000)
	for logical, fresh := range freshDocuments {
		if len(fresh) == 0 {
			t.Fatalf("fresh Documents cutover shadow migration did not load: %s", logical)
		}
		if bytes.Count(fresh, []byte("\n"))+1 > 200 {
			t.Fatalf("Documents cutover shadow migration exceeds 200 lines: %s", logical)
		}
		if _, err := postgresv1.InspectMigrationSource(fresh); err != nil {
			t.Fatalf("inspect Documents cutover shadow migration %s: %v", logical, err)
		}
		combined = append(combined, fresh...)
	}
	source := string(combined)
	for _, marker := range []string{
		"platform.cutover_job_bindings",
		"platform_async_enqueue_documents_cutover_copy_v1",
		"platform_async_assert_cutover_shadow_authority_v1",
		"'cutover_shadow'",
		"SESSION_USER='harpy_adopter'",
		"ENABLE ALWAYS TRIGGER",
		"GRANT SELECT ON TABLE platform.cutover_job_bindings TO harpy_runtime",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("Documents cutover shadow migrations lack %q", marker)
		}
	}
	if strings.Contains(source, "GRANT EXECUTE ON FUNCTION "+
		"platform.platform_async_enqueue_documents_cutover_copy_v1") {
		t.Fatal("Documents platform enqueue must not be directly executable")
	}
}
