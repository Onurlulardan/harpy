//go:build integration

package migrator

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	platformmigration "github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func currentRunnerIntegrationPlatformCatalog(
	t *testing.T,
	repositoryRoot string,
	origin postgresv1.DatabaseOrigin,
) (map[string][]byte, []postgresv1.SelectedMigration) {
	t.Helper()
	registry, err := platformmigration.LoadReleaseRegistry()
	if err != nil {
		t.Fatalf("load migration release registry: %v", err)
	}
	var bestContents map[string][]byte
	var bestSelected []postgresv1.SelectedMigration
	bestEntryCount := -1
	bestSupersessionCount := -1
	for _, digest := range registry.Digests() {
		artifact, err := registry.MigrationArtifact(digest)
		if err != nil {
			t.Fatalf("load migration manifest %s: %v", digest, err)
		}
		manifest := artifact.Value()
		selected, err := manifest.Select(origin)
		if err != nil {
			t.Fatalf("select %s migration manifest %s: %v", origin, digest, err)
		}
		contents, platformSelected, current := readCurrentPlatformSources(
			repositoryRoot, selected,
		)
		if !current || len(manifest.Entries) < bestEntryCount {
			continue
		}
		supersessionCount := selectedSupersessionCount(platformSelected)
		if len(manifest.Entries) == bestEntryCount &&
			supersessionCount <= bestSupersessionCount {
			continue
		}
		bestEntryCount = len(manifest.Entries)
		bestSupersessionCount = supersessionCount
		bestContents = contents
		bestSelected = platformSelected
	}
	if len(bestSelected) == 0 {
		t.Fatal("no embedded migration manifest matches current platform sources")
	}
	return bestContents, bestSelected
}

func readCurrentPlatformSources(
	repositoryRoot string,
	selected []postgresv1.SelectedMigration,
) (map[string][]byte, []postgresv1.SelectedMigration, bool) {
	contents := map[string][]byte{}
	platformSelected := make([]postgresv1.SelectedMigration, 0)
	for _, candidate := range selected {
		if candidate.Scope != "platform" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(
			repositoryRoot, filepath.FromSlash(candidate.Path),
		))
		if err != nil || migrationContentSHA256(content) != candidate.SHA256 {
			return nil, nil, false
		}
		contents[candidate.Path] = content
		platformSelected = append(platformSelected, candidate)
	}
	return contents, platformSelected, true
}

func selectedSupersessionCount(selected []postgresv1.SelectedMigration) int {
	count := 0
	for _, candidate := range selected {
		if strings.Contains(candidate.Path, "/supersessions/") {
			count++
		}
	}
	return count
}

func migrationContentSHA256(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}
