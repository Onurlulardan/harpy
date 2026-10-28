package migrationstreamcatalog

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoadReturnsStrictOwnerCatalogAndSortedStreams(t *testing.T) {
	root := t.TempDir()
	alpha := validStream("alpha", 20)
	alpha.Scope = "tenant"
	alpha.Phase = "expand"
	alpha.TargetDatabaseGeneration = 2
	zeta := validStream("zeta", 10)
	zeta.TargetDatabaseGeneration = 2
	writeRoot(t, root, catalogVersion, []string{streamPath("alpha"), streamPath("zeta")})
	writeOwner(t, root, ownerDocumentVersion, "alpha", []Stream{alpha})
	writeOwner(t, root, ownerDocumentVersion, "zeta", []Stream{zeta})

	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	want := Manifest{Version: catalogVersion, Owners: []Owner{
		{Name: "alpha", Streams: []Stream{alpha}},
		{Name: "zeta", Streams: []Stream{zeta}},
	}}
	if !reflect.DeepEqual(manifest, want) {
		t.Fatalf("manifest = %#v, want %#v", manifest, want)
	}
	streams := manifest.Streams()
	if len(streams) != 2 || streams[0].VersionFrom != 10 || streams[1].VersionFrom != 20 {
		t.Fatalf("globally sorted streams = %#v", streams)
	}
}

func TestLoadHasNoV1FlatFallback(t *testing.T) {
	root := t.TempDir()
	writeRawFile(t, root, rootDocumentPath, []byte(`{"version":1,"streams":[]}`))
	_, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "exact schema key") {
		t.Fatalf("v1 catalog error = %v", err)
	}
}

func TestLoadAllowsExplicitEmptyRoot(t *testing.T) {
	root := t.TempDir()
	writeRoot(t, root, catalogVersion, []string{})
	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Version != catalogVersion || manifest.Owners == nil ||
		len(manifest.Owners) != 0 || manifest.Streams() == nil {
		t.Fatalf("empty manifest = %#v", manifest)
	}
}
