package cursor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func mandatoryPredicateDigest(key codecKey, canonical []byte) [sha256.Size]byte {
	digest := hmac.New(sha256.New, key.digestKey[:])
	_, _ = digest.Write(canonical)
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func encodeDigest(digest [sha256.Size]byte) string {
	return hex.EncodeToString(digest[:])
}

func decodeDigest(value string) ([sha256.Size]byte, bool) {
	if len(value) != sha256.Size*2 {
		return [sha256.Size]byte{}, false
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || hex.EncodeToString(decoded) != value || len(decoded) != sha256.Size {
		clear(decoded)
		return [sha256.Size]byte{}, false
	}
	var digest [sha256.Size]byte
	copy(digest[:], decoded)
	clear(decoded)
	return digest, digest != [sha256.Size]byte{}
}
