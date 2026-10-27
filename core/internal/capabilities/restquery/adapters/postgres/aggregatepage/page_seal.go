package aggregatepage

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

const pageSealDomain = "harpy.restquery.aggregate-page.v1"

func cloneGroups(groups [][]byte) [][]byte {
	if groups == nil {
		return nil
	}
	result := make([][]byte, len(groups))
	for index, group := range groups {
		result[index] = append([]byte(nil), group...)
	}
	return result
}

func pageSeal(page Page) [sha256.Size]byte {
	digest := sha256.New()
	writeSealBytes(digest, []byte(pageSealDomain))
	writeSealInt(digest, int64(page.limit))
	writeSealInt(digest, int64(len(page.groups)))
	for _, group := range page.groups {
		writeSealBytes(digest, group)
	}
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writeSealBytes(digest hash.Hash, value []byte) {
	writeSealInt(digest, int64(len(value)))
	_, _ = digest.Write(value)
}

func writeSealInt(digest hash.Hash, value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	_, _ = digest.Write(encoded[:])
}
