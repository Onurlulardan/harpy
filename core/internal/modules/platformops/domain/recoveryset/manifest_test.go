package recoveryset

import (
	"bytes"
	"testing"
)

func TestManifestBindsEveryImmutableComponentIdentity(t *testing.T) {
	manifest := restoreFixture().Manifest
	document, err := manifest.CanonicalBytes()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseManifest(document)
	if err != nil || parsed.Components.DatabaseArchive.ObjectVersionID == "" {
		t.Fatalf("parse manifest: %v", err)
	}
	tampered := bytes.Replace(
		document,
		[]byte(`"object_version_id":"version-database"`),
		[]byte(`"object_version_id":"version-attacker"`),
		1,
	)
	if _, err = ParseManifest(tampered); err != nil {
		t.Fatalf("canonical component identity should remain parseable: %v", err)
	}
	if bytes.Equal(document, tampered) {
		t.Fatal("component version was not present in canonical manifest")
	}
}

func TestManifestRejectsMissingOrDigestDivergentComponent(t *testing.T) {
	missing := restoreFixture().Manifest
	missing.Components.DocumentsInventory = ComponentIdentity{}
	if _, err := missing.CanonicalBytes(); err == nil {
		t.Fatal("missing documents component was accepted")
	}
	divergent := restoreFixture().Manifest
	divergent.Components.DatabaseArchive.CiphertextSHA256 = digestFixture("9")
	if _, err := divergent.CanonicalBytes(); err == nil {
		t.Fatal("divergent database component digest was accepted")
	}
}
