package cursor

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func encodedFixture(t testing.TB) (Codec, Context, string) {
	t.Helper()
	context := buildContext(t, baseContextFixture(t))
	codec, err := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{11}, nonceBytes)),
		func() time.Time { return testNow },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := codec.Encode(
		context, []LastValue{scalarLast(t, api.LogicalTypeText, "record-a")},
	)
	if err != nil {
		t.Fatal(err)
	}
	return codec, context, token
}

func splitToken(t testing.TB, token string) []string {
	t.Helper()
	segments := strings.Split(token, ".")
	if len(segments) != cursorSegments {
		t.Fatalf("invalid test token: %q", token)
	}
	return segments
}

func joinToken(segments []string) string {
	return strings.Join(segments, ".")
}

func resealPlaintext(
	t testing.TB, key codecKey, plaintext []byte, nonceByte byte,
) string {
	t.Helper()
	nonce := make([]byte, nonceBytes)
	for index := range nonce {
		nonce[index] = nonceByte
	}
	aad := []byte(cursorVersion + "." + key.kid)
	ciphertext := key.aead.Seal(nil, nonce, plaintext, aad)
	return strings.Join([]string{
		cursorVersion, key.kid,
		base64.RawURLEncoding.EncodeToString(nonce),
		base64.RawURLEncoding.EncodeToString(ciphertext),
	}, ".")
}
