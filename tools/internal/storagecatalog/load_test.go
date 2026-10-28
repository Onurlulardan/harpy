package storagecatalog

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoadReturnsStrictMergedCatalog(t *testing.T) {
	root := t.TempDir()
	alpha := validOwner("alpha")
	alpha.Scopes = []string{"platform", "tenant"}
	alpha.MigrationDirs = []DirectoryClaim{{Path: "alpha/migrations", Scope: "tenant"}}
	alpha.StaticSQLDirs = []DirectoryClaim{{Path: "alpha/sql", Scope: "platform"}}
	alpha.PlatformSchemas = []string{"platform"}
	alpha.SchemaUsageGrants = []SchemaGrant{{
		Schema: "platform", Role: "runtime", Contract: "ADR-1",
	}}
	alpha.PlatformRoutines = []string{"platform.alpha_run"}
	alpha.RoutineExecGrants = []RoutineGrant{{
		Routine: "platform.alpha_run", IdentityArguments: "uuid",
		Role: "runtime", Contract: "ADR-1",
	}}
	alpha.PlatformTables = []string{"platform.alpha"}
	alpha.ExternalAccesses = []ExternalAccess{{
		Access: "read", TargetOwner: "zeta", Target: "platform.zeta",
		Contract: "ADR-1",
	}}
	alpha.ForeignKeys = []ForeignKey{{
		Constraint: "alpha_zeta_fk", Source: "platform.alpha",
		SourceColumns: []string{"zeta_id"}, TargetOwner: "zeta",
		Target: "platform.zeta", TargetColumns: []string{"id"},
		MatchType: "simple", OnUpdate: "no_action", OnDelete: "restrict",
		Contract: "ADR-1",
	}}
	alpha.LegacyTables = []LegacyTable{{
		Name: "public.alpha", Owner: "alpha", Adapter: "alpha.go",
		CutoverGate: "alpha/v2_only",
	}}
	alpha.DynamicNamespaces = []DynamicNamespaceClaim{{
		Kind: "table", Prefix: "alpha_dynamic_", Authority: "alpha.compiler",
	}}
	zeta := validOwner("zeta")
	writeRoot(t, root, catalogVersion, []string{ownerPath("alpha"), ownerPath("zeta")})
	writeOwner(t, root, ownerDocumentVersion, alpha)
	writeOwner(t, root, ownerDocumentVersion, zeta)

	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	want := Manifest{Version: catalogVersion, Owners: []Owner{alpha, zeta}}
	if !reflect.DeepEqual(manifest, want) {
		t.Fatalf("manifest = %#v, want %#v", manifest, want)
	}
}

func TestLoadHasNoV1Fallback(t *testing.T) {
	root := t.TempDir()
	writeRawFile(t, root, rootDocumentPath, []byte(`{"version":1,"owners":[]}`))
	_, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "exact schema key") {
		t.Fatalf("v1 catalog error = %v", err)
	}
}

func TestLoadAllowsExplicitEmptyCatalog(t *testing.T) {
	root := t.TempDir()
	writeRoot(t, root, catalogVersion, []string{})
	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Version != catalogVersion || manifest.Owners == nil ||
		len(manifest.Owners) != 0 {
		t.Fatalf("empty manifest = %#v", manifest)
	}
}
