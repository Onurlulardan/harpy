package schemacontract

import "testing"

func TestFreshExtensionManifestDigestIsVerified(t *testing.T) {
	digest, err := freshExtensionDigest()
	if err != nil {
		t.Fatal(err)
	}
	if len(digest) != 64 {
		t.Fatalf("extension digest = %s", digest)
	}
	again, err := freshExtensionDigest()
	if err != nil || again != digest {
		t.Fatalf("extension digest changed: %s %v", again, err)
	}
}
