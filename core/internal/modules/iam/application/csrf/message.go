package csrf

import (
	"crypto/hmac"
	"crypto/sha256"
)

func authenticationCode(
	key []byte,
	keyID, nonce, selector, origin, issuedAt, expiry string,
) [sha256.Size]byte {
	hash := hmac.New(sha256.New, key)
	writePart(hash, "harpy-csrf-v1")
	writePart(hash, keyID)
	writePart(hash, nonce)
	writePart(hash, selector)
	writePart(hash, origin)
	writePart(hash, issuedAt)
	writePart(hash, expiry)
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

type byteWriter interface{ Write([]byte) (int, error) }

func writePart(writer byteWriter, value string) {
	_, _ = writer.Write([]byte{0})
	_, _ = writer.Write([]byte(value))
}
