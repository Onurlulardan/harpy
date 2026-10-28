package migrationmanifest

import (
	"path/filepath"
	"strings"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestValidateReleaseFrontierAdvanceAllowsExactSupersession(t *testing.T) {
	directory := t.TempDir()
	previous := releaseFrontierFixture(12)
	writeReleaseFrontier(t, directory, previous)

	current := exactSupersessionFrontier(previous, "b")
	if err := ValidateReleaseFrontierAdvance(directory, current); err != nil {
		t.Fatalf("exact supersession was rejected: %v", err)
	}
	writeReleaseFrontier(t, directory, current)
	if err := ValidateReleaseFrontierAdvance(directory, current); err != nil {
		t.Fatalf("published supersession lineage was rejected: %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceRejectsInPlaceDigestChange(t *testing.T) {
	directory := t.TempDir()
	previous := releaseFrontierFixture(12)
	writeReleaseFrontier(t, directory, previous)

	current := previous
	current.Entries = append([]postgresv1.MigrationEntry(nil), previous.Entries...)
	current.Entries[0].Variants.Fresh.SHA256 = strings.Repeat("b", 64)
	err := ValidateReleaseFrontierAdvance(directory, current)
	if err == nil || !strings.Contains(err.Error(), "does not append") {
		t.Fatalf("in-place digest change error = %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceAllowsSecondLevelSupersession(t *testing.T) {
	directory := t.TempDir()
	original := releaseFrontierFixture(12)
	writeReleaseFrontier(t, directory, original)
	first := exactSupersessionFrontier(original, "b")
	writeReleaseFrontier(t, directory, first)

	second := exactSecondLevelSupersessionFrontier(first, "c")
	if err := ValidateReleaseFrontierAdvance(directory, second); err != nil {
		t.Fatalf("second-level supersession was rejected: %v", err)
	}
	writeReleaseFrontier(t, directory, second)
	if err := ValidateReleaseFrontierAdvance(directory, second); err != nil {
		t.Fatalf("published second-level lineage was rejected: %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceAllowsThirdLevelSupersession(t *testing.T) {
	directory := t.TempDir()
	original := releaseFrontierFixture(12)
	first := exactSupersessionFrontier(original, "b")
	second := exactSecondLevelSupersessionFrontier(first, "c")
	writeReleaseFrontier(t, directory, original)
	writeReleaseFrontier(t, directory, first)
	writeReleaseFrontier(t, directory, second)

	third := exactThirdLevelSupersessionFrontier(second, "d")
	if err := ValidateReleaseFrontierAdvance(directory, third); err != nil {
		t.Fatalf("third-level supersession was rejected: %v", err)
	}
	writeReleaseFrontier(t, directory, third)
	if err := ValidateReleaseFrontierAdvance(directory, third); err != nil {
		t.Fatalf("published third-level lineage was rejected: %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceRejectsFourthLevelSupersession(t *testing.T) {
	directory := t.TempDir()
	original := releaseFrontierFixture(12)
	first := exactSupersessionFrontier(original, "b")
	second := exactSecondLevelSupersessionFrontier(first, "c")
	third := exactThirdLevelSupersessionFrontier(second, "d")
	for _, release := range []postgresv1.MigrationManifest{original, first, second, third} {
		writeReleaseFrontier(t, directory, release)
	}
	fourth := moveSupersessionFrontier(third, "v4", "e")
	if err := ValidateReleaseFrontierAdvance(directory, fourth); err == nil {
		t.Fatal("fourth-level supersession was accepted")
	}
}

func exactSupersessionFrontier(
	previous postgresv1.MigrationManifest,
	digestCharacter string,
) postgresv1.MigrationManifest {
	current := previous
	current.Entries = append([]postgresv1.MigrationEntry(nil), previous.Entries...)
	entry := current.Entries[0]
	entry.Variants.Fresh = &postgresv1.MigrationVariant{
		Path: "core/internal/modules/documents/migrations/platform/supersessions/fresh/" +
			filepath.Base(previous.Entries[0].Variants.Fresh.Path),
		SHA256: strings.Repeat(digestCharacter, 64),
	}
	entry.Variants.V1Adopted = &postgresv1.MigrationVariant{
		Path: "core/internal/modules/documents/migrations/platform/supersessions/v1_adopted/" +
			filepath.Base(previous.Entries[0].Variants.V1Adopted.Path),
		SHA256: strings.Repeat(digestCharacter, 64),
	}
	current.Entries[0] = entry
	return current
}

func exactSecondLevelSupersessionFrontier(
	previous postgresv1.MigrationManifest,
	digestCharacter string,
) postgresv1.MigrationManifest {
	return moveSupersessionFrontier(previous, "v2", digestCharacter)
}

func exactThirdLevelSupersessionFrontier(
	previous postgresv1.MigrationManifest,
	digestCharacter string,
) postgresv1.MigrationManifest {
	return moveSupersessionFrontier(previous, "v3", digestCharacter)
}

func moveSupersessionFrontier(
	previous postgresv1.MigrationManifest,
	level, digestCharacter string,
) postgresv1.MigrationManifest {
	current := previous
	current.Entries = append([]postgresv1.MigrationEntry(nil), previous.Entries...)
	entry := current.Entries[0]
	entry.Variants.Fresh = &postgresv1.MigrationVariant{
		Path: "core/internal/modules/documents/migrations/platform/supersessions/" +
			level + "/fresh/" +
			filepath.Base(previous.Entries[0].Variants.Fresh.Path),
		SHA256: strings.Repeat(digestCharacter, 64),
	}
	entry.Variants.V1Adopted = &postgresv1.MigrationVariant{
		Path: "core/internal/modules/documents/migrations/platform/supersessions/" +
			level + "/v1_adopted/" +
			filepath.Base(previous.Entries[0].Variants.V1Adopted.Path),
		SHA256: strings.Repeat(digestCharacter, 64),
	}
	current.Entries[0] = entry
	return current
}
