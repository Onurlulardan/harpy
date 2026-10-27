package migration

import (
	"strings"
	"testing"
)

func TestSelectFreshDocumentsKeepsTerminalSupersession(t *testing.T) {
	const version = "20260730120104_repair_installation_layout_publications_v1.fresh.sql"
	root := "core/internal/modules/system_resource_layout/migrations/tenant/"
	paths := []string{
		root + "fresh/" + version,
		root + "supersessions/fresh/" + version,
		root + "supersessions/v2/fresh/" + version,
		root + "supersessions/v3/fresh/" + version,
	}
	documents := make([]Document, len(paths))
	for index, documentPath := range paths {
		documents[index] = freshTestDocument(t, documentPath, index)
	}
	selected, err := SelectFreshDocuments(documents)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || !strings.Contains(selected[0].Path(), "/supersessions/v3/fresh/") {
		t.Fatalf("terminal supersession = %#v", selected)
	}
}

func TestSelectFreshDocumentsRejectsAdoptedAndAmbiguousCopies(t *testing.T) {
	adopted := freshTestDocument(
		t, "core/internal/platform/migrator/migrations/v1_adopted/"+
			"20260713010000_create_foundation.v1_adopted.sql", 0,
	)
	if _, err := SelectFreshDocuments([]Document{adopted}); err == nil ||
		!strings.Contains(err.Error(), "v1_adopted is not present") {
		t.Fatalf("adopted source error = %v", err)
	}
	first := freshTestDocument(t, testFreshPath, 1)
	second := freshTestDocument(
		t, "core/internal/platform/migrator/migrations/fresh/"+
			"20260713010000_create_foundation_again.fresh.sql", 2,
	)
	if _, err := SelectFreshDocuments([]Document{first, second}); err == nil {
		t.Fatal("ambiguous fresh supersession was accepted")
	}
}

func freshTestDocument(t *testing.T, documentPath string, salt int) Document {
	t.Helper()
	body := "-- +goose Up\nSELECT " + string(rune('1'+salt)) + ";\n"
	document, err := NewDocument("platform_migrator", "platform", documentPath, []byte(body))
	if err != nil {
		t.Fatal(err)
	}
	return document
}
