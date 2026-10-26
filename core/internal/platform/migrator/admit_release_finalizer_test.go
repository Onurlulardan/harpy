package migrator

import (
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestAdmitReleaseFinalizerKeepsCatalogPlatformSelection(t *testing.T) {
	t.Parallel()
	finalizer := &releaseFinalizerFixture{}
	selected := []postgresv1.SelectedMigration{
		{Owner: "platform_cutover", Scope: platformMigrationScope},
		{Owner: "catalog", Scope: platformMigrationScope, Version: 20260718180000},
	}
	if admitReleaseFinalizer(selected, finalizer) != finalizer {
		t.Fatal("catalog platform selection dropped the release finalizer")
	}
}

func TestAdmitReleaseFinalizerSkipsFoundationSelection(t *testing.T) {
	t.Parallel()
	finalizer := &releaseFinalizerFixture{}
	selected := []postgresv1.SelectedMigration{
		{Owner: "platform_cutover", Scope: platformMigrationScope, Version: 20260714014100},
		{Owner: "catalog", Scope: tenantMigrationScope, Version: 20260718180100},
	}
	if admitReleaseFinalizer(selected, finalizer) != nil {
		t.Fatal("foundation selection kept the catalog release finalizer")
	}
}

func TestAdmitReleaseFinalizerAllowsNilFinalizer(t *testing.T) {
	t.Parallel()
	selected := []postgresv1.SelectedMigration{
		{Owner: "catalog", Scope: platformMigrationScope},
	}
	if admitReleaseFinalizer(selected, nil) != nil {
		t.Fatal("nil release finalizer was reconstructed")
	}
}
