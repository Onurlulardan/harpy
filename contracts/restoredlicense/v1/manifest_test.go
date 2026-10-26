package restoredlicensev1

import "testing"

func TestManifestIsOneRestoreBoundCommand(t *testing.T) {
	manifest := Manifest()
	if len(manifest) != 1 || manifest[0] != ReconcileRestoredLicense ||
		len(ManifestSHA256()) != 64 {
		t.Fatalf("restored license manifest drifted: %#v", manifest)
	}
	manifest[0] = "mutated"
	if Manifest()[0] != ReconcileRestoredLicense {
		t.Fatal("manifest returned shared mutable storage")
	}
}
