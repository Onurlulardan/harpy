package cursor

import (
	"bytes"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func TestMandatoryPredicateDigestUsesDedicatedExactHKDFContext(t *testing.T) {
	codec, context, token := encodedFixture(t)
	plaintext, _, err := codec.decryptToken(token)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(plaintext)
	var claims wireClaims
	if err := attestation.DecodeStrict(plaintext, &claims); err != nil {
		t.Fatal(err)
	}
	derived, err := hkdf.Key(
		sha256.New, bytes.Repeat([]byte{1}, 32), nil,
		"harpy/rest-query/cursor-mandatory-predicate/v1", sha256.Size,
	)
	if err != nil {
		t.Fatal(err)
	}
	digest := hmac.New(sha256.New, derived)
	_, _ = digest.Write(context.mandatoryPredicate)
	expected := hex.EncodeToString(digest.Sum(nil))
	clear(derived)
	if claims.MandatoryPredicateHMACSHA256 != expected {
		t.Fatal("mandatory predicate digest did not use the dedicated HKDF context")
	}
	direct := sha256.Sum256(context.mandatoryPredicate)
	if claims.MandatoryPredicateHMACSHA256 == hex.EncodeToString(direct[:]) {
		t.Fatal("mandatory predicate used an unkeyed digest")
	}
}

func TestClaimDigestsRequireCanonicalLowercaseNonzeroSHA256(t *testing.T) {
	valid := strings.Repeat("ab", sha256.Size)
	if _, ok := decodeDigest(valid); !ok {
		t.Fatal("canonical digest rejected")
	}
	for _, invalid := range []string{
		strings.ToUpper(valid), strings.Repeat("0", sha256.Size*2), valid[:len(valid)-1],
	} {
		if _, ok := decodeDigest(invalid); ok {
			t.Fatalf("invalid digest accepted: %q", invalid)
		}
	}
}
