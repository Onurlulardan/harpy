package executor

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

const resultSealDomain = "harpy.restquery.postgres.execution-result.v1"

func resultSeal(result Result) ([sha256.Size]byte, error) {
	records, recordsErr := result.page.Records()
	hasMore, moreErr := result.page.HasMore()
	requestedSize, requestedErr := result.page.RequestedSize()
	nextCursor, present, cursorErr := result.page.NextCursor()
	if recordsErr != nil || moreErr != nil || requestedErr != nil || cursorErr != nil ||
		hasMore != present {
		return [sha256.Size]byte{}, ErrContract
	}
	digest := sha256.New()
	sealBytes(digest, []byte(resultSealDomain))
	sealBool(digest, result.hasCount)
	sealInt(digest, result.exactCount)
	sealBool(digest, hasMore)
	sealInt(digest, int64(requestedSize))
	sealBytes(digest, []byte(nextCursor))
	sealInt(digest, int64(len(records)))
	for _, record := range records {
		sealBytes(digest, record)
	}
	var seal [sha256.Size]byte
	copy(seal[:], digest.Sum(nil))
	return seal, nil
}

func sealBytes(digest hash.Hash, value []byte) {
	sealInt(digest, int64(len(value)))
	_, _ = digest.Write(value)
}

func sealInt(digest hash.Hash, value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	_, _ = digest.Write(encoded[:])
}

func sealBool(digest hash.Hash, value bool) {
	encoded := byte(0)
	if value {
		encoded = 1
	}
	_, _ = digest.Write([]byte{encoded})
}
