package authentication

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestHMACKeyIDRequiresCanonicalRawURLBase64(t *testing.T) {
	canonical := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0xfb}, 12))
	if !validHMACKeyID(canonical) || !strings.ContainsAny(canonical, "-_ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		t.Fatalf("canonical mixed-case key ID rejected: %q", canonical)
	}
	for _, value := range []string{
		"not+base64url-value", canonical + "=",
		base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, 11)),
	} {
		if validHMACKeyID(value) {
			t.Fatalf("invalid key ID accepted: %q", value)
		}
	}
}

func TestKMSKeyIDRequiresCanonicalGoogleCloudKeyVersion(t *testing.T) {
	valid := "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7"
	if !validKMSKeyID(valid) {
		t.Fatal("valid KMS key ID rejected")
	}
	for _, value := range []string{
		"", "kms://iam/totp/current", valid + "/", strings.Repeat("x", 513),
	} {
		if validKMSKeyID(value) {
			t.Fatalf("invalid KMS key ID accepted: %q", value)
		}
	}
}
