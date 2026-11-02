package schemacontract

import (
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

const foundationMigrationDigest = "f9fef3e093296fb402caed0a48415560aff29224cdbc62b56926f1209718b266"

func TestLoadStartupKeepsFreshRegistryWhenDigestMatches(t *testing.T) {
	document := startupSQLDocument(t)
	documents := []migration.Document{document}
	fresh, err := migration.LoadStartupRegistry(documents)
	if err != nil {
		t.Fatal(err)
	}
	digest := fresh.Digests()[0]
	artifact := contractArtifact(t, digest, "extension-fresh")
	consulted := false
	embedded, registry, err := loadStartup(
		documents,
		func() (binarycontract.Artifact, error) { return artifact, nil },
		func() (*migration.ReleaseRegistry, error) {
			consulted = true
			return nil, errReleaseConsulted
		},
	)
	if err != nil || consulted || !embedded.IsVerified() || registry == nil ||
		registry.FreshEntryCount() != 1 || registry.Digests()[0] != digest {
		t.Fatalf("fresh startup consulted=%v registry=%#v err=%v", consulted, registry, err)
	}
}

func TestLoadStartupRejectsUnknownDigest(t *testing.T) {
	document := startupSQLDocument(t)
	artifact := contractArtifact(t, strings.Repeat("ab", 32), "extension-unknown")
	embedded, registry, err := loadStartup(
		[]migration.Document{document},
		func() (binarycontract.Artifact, error) { return artifact, nil },
		migration.LoadReleaseRegistry,
	)
	if err == nil || registry != nil || embedded.IsVerified() ||
		!strings.Contains(err.Error(), "migration manifest digest is not embedded") {
		t.Fatalf("unknown digest startup = %#v, %v", registry, err)
	}
}

func startupSQLDocument(t *testing.T) migration.Document {
	t.Helper()
	document, err := migration.NewDocument(
		"platform_migrator", "platform",
		"core/internal/platform/migrator/migrations/fresh/"+
			"20260713010000_create_foundation.fresh.sql",
		[]byte("-- +goose Up\nSELECT 1;\n"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

type releaseConsultedError struct{}

func (releaseConsultedError) Error() string {
	return "release archive must not be consulted"
}

var errReleaseConsulted = releaseConsultedError{}
