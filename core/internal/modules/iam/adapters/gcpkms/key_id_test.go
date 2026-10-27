package gcpkms

import "testing"

const (
	activeFixtureKey = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7"
	retiredFixtureKey = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/6"
)

func TestParseKeyVersionAcceptsCanonicalVersionAndBindsCryptoKey(t *testing.T) {
	active, err := parseKeyVersion(activeFixtureKey)
	if err != nil {
		t.Fatalf("parse active key: %v", err)
	}
	retired, err := parseKeyVersion(retiredFixtureKey)
	if err != nil {
		t.Fatalf("parse retired key: %v", err)
	}
	resource, _ := active.Resource()
	if resource != "projects/harpy-prod/locations/europe-west1/"+
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7" {
		t.Fatalf("unexpected resource %q", resource)
	}
	if !active.SameCryptoKey(retired) {
		t.Fatal("versions of the same crypto key were not accepted")
	}
}

func TestParseKeyVersionRejectsNoncanonicalOrUnboundedValues(t *testing.T) {
	values := []string{
		"", "kms://iam/totp/current", "gcp-kms://projects/x/locations/global/" +
			"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/1",
		activeFixtureKey + "/", activeFixtureKey + "?x=1",
		"gcp-kms://projects/harpy-prod/locations/global/keyRings/iam/" +
			"cryptoKeys/other/cryptoKeyVersions/0",
	}
	for _, value := range values {
		if _, err := parseKeyVersion(value); err == nil {
			t.Fatalf("invalid key ID accepted: %q", value)
		}
	}
}
