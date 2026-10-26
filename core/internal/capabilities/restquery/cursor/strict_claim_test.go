package cursor

import (
	"bytes"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestDecodeRequiresStrictExactCanonicalVersionedClaims(t *testing.T) {
	codec, context, token := encodedFixture(t)
	plaintext, key, err := codec.decryptToken(token)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(plaintext)
	mutations := [][]byte{
		bytes.Replace(plaintext, []byte("{"), []byte(`{"unknown":true,`), 1),
		bytes.Replace(plaintext, []byte(`"version":"v1"`), []byte(`"version":"v1","version":"v1"`), 1),
		bytes.Replace(plaintext, []byte(`,"version":"v1"`), nil, 1),
		append(append([]byte(nil), plaintext...), '\n'),
		append(append([]byte(nil), plaintext...), []byte(`{}`)...),
		bytes.Replace(plaintext, []byte(`"version":"v1"`), []byte(`"version":"v2"`), 1),
		bytes.Replace(plaintext, []byte(`"value":"record-a"`), []byte(`"value":true`), 1),
		bytes.Replace(plaintext, []byte(`"direction":"asc"`), []byte(`"direction":"asc","direction":"asc"`), 1),
	}
	for index, mutation := range mutations {
		if bytes.Equal(mutation, plaintext) {
			t.Fatalf("mutation %d did not change plaintext", index)
		}
		candidate := resealPlaintext(t, key, mutation, byte(index+20))
		_, decodeErr := codec.Decode(candidate, context)
		requireQueryCode(t, decodeErr, restqueryv1.ErrorQueryCursorInvalid)
	}
}
