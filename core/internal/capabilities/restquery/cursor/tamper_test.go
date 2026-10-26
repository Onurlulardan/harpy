package cursor

import (
	"encoding/base64"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestDecodeRejectsMalformedWireBase64AADAndTamper(t *testing.T) {
	codec, context, token := encodedFixture(t)
	if _, err := codec.Decode(token, Context{}); err == nil {
		t.Fatal("zero expected context accepted")
	} else {
		requireQueryCode(t, err, restqueryv1.ErrorQueryCursorInvalid)
	}
	segments := splitToken(t, token)
	tamperedCiphertext := append([]string(nil), segments...)
	last := len(tamperedCiphertext[3]) - 1
	if tamperedCiphertext[3][last] == 'A' {
		tamperedCiphertext[3] = tamperedCiphertext[3][:last] + "B"
	} else {
		tamperedCiphertext[3] = tamperedCiphertext[3][:last] + "A"
	}
	wrongAAD := append([]string(nil), segments...)
	wrongAAD[1] = "previous"
	badPadding := append([]string(nil), segments...)
	badPadding[2] += "="
	shortNonce := append([]string(nil), segments...)
	shortNonce[2] = base64.RawURLEncoding.EncodeToString(make([]byte, nonceBytes-1))
	shortCiphertext := append([]string(nil), segments...)
	shortCiphertext[3] = base64.RawURLEncoding.EncodeToString(make([]byte, 15))
	tests := []string{
		"", "v1", "v1.current.a.b.extra", "v2." + strings.Join(segments[1:], "."),
		joinToken(tamperedCiphertext), joinToken(wrongAAD), joinToken(badPadding),
		joinToken(shortNonce), joinToken(shortCiphertext),
		strings.Repeat("a", maximumCursorBytes+1),
	}
	for _, candidate := range tests {
		_, err := codec.Decode(candidate, context)
		requireQueryCode(t, err, restqueryv1.ErrorQueryCursorInvalid)
		if strings.Contains(err.Error(), candidate) && candidate != "" {
			t.Fatal("cursor error leaked attacker-controlled token")
		}
	}
}

func TestEncryptedPlaintextCarriesNoRawMandatoryPredicate(t *testing.T) {
	codec, _, token := encodedFixture(t)
	plaintext, _, err := codec.decryptToken(token)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(plaintext)
	for _, prohibited := range []string{testTenantID, `"active":true`, "record-a"} {
		if strings.Contains(token, prohibited) {
			t.Fatalf("wire exposed %q", prohibited)
		}
	}
	if strings.Contains(string(plaintext), `"active":true`) {
		t.Fatal("plaintext exposed raw mandatory predicate")
	}
	if strings.Contains(string(plaintext), "canonical_bound") ||
		strings.Contains(string(plaintext), `"bound"`) {
		t.Fatal("plaintext carried catalog canonical bounds")
	}
	if !strings.Contains(string(plaintext), `"mandatory_predicate_hmac_sha256"`) {
		t.Fatal("plaintext omitted mandatory predicate digest")
	}
}

func TestBase64URLDecoderRejectsPaddingAndNonzeroTrailingBits(t *testing.T) {
	if decoded, ok := decodeCanonicalBase64URL("_w"); !ok || len(decoded) != 1 || decoded[0] != 0xff {
		t.Fatal("canonical raw base64url rejected")
	}
	for _, invalid := range []string{"_w=", "_x", "+w", "/w"} {
		if _, ok := decodeCanonicalBase64URL(invalid); ok {
			t.Fatalf("non-canonical base64url accepted: %q", invalid)
		}
	}
}
