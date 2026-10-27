package credentialdeliveryv1

import "testing"

func TestManifestIsExactOrderedCredentialRecovery(t *testing.T) {
	manifest := Manifest()
	if len(manifest) != 2 || manifest[0] != SupersedeTemporaryCredential ||
		manifest[1] != RecordDeliveryAttempt || len(ManifestSHA256()) != 64 {
		t.Fatalf("credential delivery manifest drifted: %#v", manifest)
	}
	manifest[0] = "mutated"
	if Manifest()[0] != SupersedeTemporaryCredential {
		t.Fatal("manifest returned shared mutable storage")
	}
}
