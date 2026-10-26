package licensebootstrapv1

import "testing"

func TestManifestIsOnePurposeBoundCommand(t *testing.T) {
	manifest := Manifest()
	if len(manifest) != 1 || manifest[0] != PublishLicenseDerivedOffering ||
		len(ManifestSHA256()) != 64 {
		t.Fatalf("license bootstrap manifest drifted: %#v", manifest)
	}
	manifest[0] = "mutated"
	if Manifest()[0] != PublishLicenseDerivedOffering {
		t.Fatal("manifest returned shared mutable storage")
	}
}
