package migrator

import (
	"reflect"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestMergeMigrationResultsAreCanonicalUnique(t *testing.T) {
	platform := []postgresv1.SelectedMigration{{Version: 100}, {Version: 300}}
	tenant := []postgresv1.SelectedMigration{{Version: 200}, {Version: 200}}
	merged, err := mergeMigrationSelections(platform, tenant)
	if err != nil {
		t.Fatal(err)
	}
	versions := []int64{merged[0].Version, merged[1].Version, merged[2].Version}
	if !reflect.DeepEqual(versions, []int64{100, 200, 300}) {
		t.Fatalf("merged selections = %v", versions)
	}
	if applied := mergeAppliedVersions(
		[]int64{300, 100}, []int64{200, 200},
	); !reflect.DeepEqual(applied, []int64{100, 200, 300}) {
		t.Fatalf("merged applied versions = %v", applied)
	}
}
