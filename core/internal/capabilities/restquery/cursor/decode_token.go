package cursor

import (
	"encoding/base64"
	"strings"
)

func (codec Codec) decryptToken(token string) ([]byte, codecKey, error) {
	if len(token) < 1 || len(token) > maximumCursorBytes {
		return nil, codecKey{}, invalidCursorError()
	}
	segments := strings.Split(token, ".")
	if len(segments) != cursorSegments || segments[0] != cursorVersion ||
		segments[1] == "" || segments[2] == "" || segments[3] == "" {
		return nil, codecKey{}, invalidCursorError()
	}
	key, known := codec.decodeKey(segments[1])
	if !known {
		return nil, codecKey{}, invalidCursorError()
	}
	nonce, nonceOK := decodeCanonicalBase64URL(segments[2])
	if !nonceOK || len(nonce) != nonceBytes {
		clear(nonce)
		return nil, codecKey{}, invalidCursorError()
	}
	ciphertext, ciphertextOK := decodeCanonicalBase64URL(segments[3])
	if !ciphertextOK || len(ciphertext) < key.aead.Overhead() {
		clear(nonce)
		clear(ciphertext)
		return nil, codecKey{}, invalidCursorError()
	}
	aad := []byte(cursorVersion + "." + key.kid)
	plaintext, err := key.aead.Open(nil, nonce, ciphertext, aad)
	clear(nonce)
	clear(ciphertext)
	if err != nil || len(plaintext) < 1 || len(plaintext) > maximumPlaintextBytes {
		clear(plaintext)
		return nil, codecKey{}, invalidCursorError()
	}
	return plaintext, key, nil
}

func decodeCanonicalBase64URL(value string) ([]byte, bool) {
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || base64.RawURLEncoding.EncodeToString(decoded) != value {
		clear(decoded)
		return nil, false
	}
	return decoded, true
}
