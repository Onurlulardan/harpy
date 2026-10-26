package identifier

import (
	"crypto/sha256"
	"testing"
)

func TestNormalizedSHA256HashesCanonicalBytes(t *testing.T) {
	t.Parallel()
	value, err := NormalizePortalV1(" Example ")
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256([]byte("example"))
	if value.NormalizedSHA256() != want {
		t.Fatal("identifier digest does not bind canonical bytes")
	}
}

func TestZeroIdentifierValueIsInvalid(t *testing.T) {
	t.Parallel()
	if (Value{}).Valid() || (Value{}).Kind() != "" ||
		(Value{}).Display() != "" || (Value{}).Normalized() != "" {
		t.Fatal("zero identifier value must be inert")
	}
}
