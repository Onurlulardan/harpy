package migrationmanifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateReleaseFrontierAdvanceForPublishRetractsUnpublishedMaximal(t *testing.T) {
	directory := t.TempDir()
	base := releaseFrontierFixture(12)
	writeReleaseFrontier(t, directory, base)
	unpublished := appendReleaseFrontierAt(base, 20260722000001, 13, "b")
	writeReleaseFrontier(t, directory, unpublished)
	_, digest, err := Encode(unpublished)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteUnpublishedMarker(directory, digest); err != nil {
		t.Fatal(err)
	}
	replacement := appendReleaseFrontierAt(base, 20260722000001, 13, "c")
	if err := ValidateReleaseFrontierAdvanceForPublish(directory, replacement); err != nil {
		t.Fatalf("unpublished maximal rewrite was rejected: %v", err)
	}
	artifact := filepath.Join(directory, digest+immutableManifestSuffix)
	if _, err := os.Stat(artifact); !os.IsNotExist(err) {
		t.Fatal("unpublished maximal artifact was retained")
	}
	if err := ValidateReleaseFrontierAdvance(directory, replacement); err != nil {
		t.Fatalf("rewritten frontier was not accepted after retract: %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceForPublishKeepsAppliedMaximal(t *testing.T) {
	directory := t.TempDir()
	base := releaseFrontierFixture(12)
	writeReleaseFrontier(t, directory, base)
	applied := appendReleaseFrontierAt(base, 20260722000001, 13, "b")
	writeReleaseFrontier(t, directory, applied)
	replacement := appendReleaseFrontierAt(base, 20260722000001, 13, "c")
	err := ValidateReleaseFrontierAdvanceForPublish(directory, replacement)
	if err == nil || !strings.Contains(err.Error(), "maximal immutable release") {
		t.Fatalf("applied maximal rewrite error = %v", err)
	}
}

func TestValidateReleaseFrontierAdvanceDoesNotRetractUnpublished(t *testing.T) {
	directory := t.TempDir()
	base := releaseFrontierFixture(12)
	writeReleaseFrontier(t, directory, base)
	unpublished := appendReleaseFrontierAt(base, 20260722000001, 13, "b")
	writeReleaseFrontier(t, directory, unpublished)
	_, digest, err := Encode(unpublished)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteUnpublishedMarker(directory, digest); err != nil {
		t.Fatal(err)
	}
	replacement := appendReleaseFrontierAt(base, 20260722000001, 13, "c")
	err = ValidateReleaseFrontierAdvance(directory, replacement)
	if err == nil || !strings.Contains(err.Error(), "maximal immutable release") {
		t.Fatalf("read-only validate error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(directory, digest+immutableManifestSuffix)); err != nil {
		t.Fatal("read-only validate deleted unpublished maximal")
	}
}
