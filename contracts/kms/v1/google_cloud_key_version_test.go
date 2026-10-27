package kmsv1

import "testing"

const keyFixture = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
	"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7"

func TestGoogleCloudKeyVersionIsCanonicalAndVersionBound(t *testing.T) {
	version, err := ParseGoogleCloudKeyVersion(keyFixture)
	if err != nil || version.Validate() != nil {
		t.Fatalf("parse key version: %v", err)
	}
	value, _ := version.Value()
	resource, _ := version.Resource()
	if value != keyFixture || resource == keyFixture ||
		resource != "projects/harpy-prod/locations/europe-west1/"+
			"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7" {
		t.Fatalf("unexpected key version: value=%q resource=%q", value, resource)
	}
	previous, _ := ParseGoogleCloudKeyVersion(
		"gcp-kms://projects/harpy-prod/locations/europe-west1/" +
			"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/6",
	)
	if !version.SameCryptoKey(previous) {
		t.Fatal("same key rotation versions did not match")
	}
}

func TestGoogleCloudKeyVersionRejectsAliasesAndMalformedResources(t *testing.T) {
	values := []string{
		"", "kms://iam/totp/current", keyFixture + "/", keyFixture + "?x=1",
		"gcp-kms://projects/x/locations/global/keyRings/iam/cryptoKeys/totp/" +
			"cryptoKeyVersions/1",
		"gcp-kms://projects/harpy-prod/locations/global/keyRings/iam/" +
			"cryptoKeys/totp/cryptoKeyVersions/0",
	}
	for _, value := range values {
		if _, err := ParseGoogleCloudKeyVersion(value); err == nil {
			t.Fatalf("invalid key version accepted: %q", value)
		}
	}
}
