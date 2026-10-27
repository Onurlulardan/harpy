package schemacontract

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func TestLoadStartupRejectsEmptyFreshSchema(t *testing.T) {
	_, registry, err := LoadStartup(nil)
	if err == nil || registry != nil {
		t.Fatalf("startup contract = %#v, %v", registry, err)
	}
	if !strings.Contains(err.Error(), "Harpy fresh schema is not present") {
		t.Fatalf("startup error = %s", err)
	}
}

func TestLoadStartupRejectsMismatchedFreshManifest(t *testing.T) {
	document, err := migration.NewDocument(
		"platform_migrator", "platform",
		"core/internal/platform/migrator/migrations/fresh/"+
			"20260713010000_create_foundation.fresh.sql",
		[]byte("-- +goose Up\nSELECT 1;\n"),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, registry, err := LoadStartup([]migration.Document{document})
	if err == nil || registry != nil ||
		!strings.Contains(err.Error(), "migration manifest digest is not embedded") {
		t.Fatalf("mismatched startup = %#v, %v", registry, err)
	}
}
