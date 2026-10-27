package migrator

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestReleaseGenerationBoundaryAllowsTenantAtPlatformFrontier(t *testing.T) {
	platform := []postgresv1.SelectedMigration{{
		TargetDatabaseGeneration: 21, NextMinimumRuntimeGeneration: 15,
	}}
	tenant := []postgresv1.SelectedMigration{{
		TargetDatabaseGeneration: 21, NextMinimumRuntimeGeneration: 15,
	}}
	boundary, err := selectReleaseGenerationBoundary(platform, tenant)
	if err != nil {
		t.Fatalf("select release generation boundary: %v", err)
	}
	if boundary.databaseGeneration != 21 || boundary.minimumRuntimeGeneration != 15 {
		t.Fatalf("release generation boundary = %+v", boundary)
	}
}

func TestReleaseGenerationBoundaryRejectsTenantOnlyAdvance(t *testing.T) {
	platform := []postgresv1.SelectedMigration{{
		TargetDatabaseGeneration: 21, NextMinimumRuntimeGeneration: 15,
	}}
	tenant := []postgresv1.SelectedMigration{{
		TargetDatabaseGeneration: 22, NextMinimumRuntimeGeneration: 15,
	}}
	if _, err := selectReleaseGenerationBoundary(platform, tenant); err == nil {
		t.Fatal("tenant-only release generation advance was accepted")
	}
}

func TestReleaseGenerationBoundaryRejectsMissingPlatformCatalog(t *testing.T) {
	if _, err := selectReleaseGenerationBoundary(nil, nil); err == nil {
		t.Fatal("missing platform migration catalog was accepted")
	}
}
