package maintenance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

func directoryLookupHMAC(key DirectoryKey, normalized []byte) [sha256.Size]byte {
	digest := hmac.New(sha256.New, key.Material[:])
	_, _ = digest.Write(normalized)
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writeDirectoryDigest(digest hash.Hash, values ...[]byte) {
	var size [8]byte
	for _, value := range values {
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = digest.Write(size[:])
		_, _ = digest.Write(value)
	}
}

func wipeDirectoryKeys(keys []DirectoryKey) {
	for index := range keys {
		clear(keys[index].Material[:])
	}
}
