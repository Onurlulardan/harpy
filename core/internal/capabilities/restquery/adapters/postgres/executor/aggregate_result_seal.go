package executor

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

const aggregateResultSealDomain = "forgecrud.restquery.postgres.aggregate-result.v1"

func aggregateResultSeal(result AggregateResult) ([sha256.Size]byte, error) {
	groups, groupsErr := result.page.Groups()
	limit, limitErr := result.page.Limit()
	if groupsErr != nil || limitErr != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	digest := sha256.New()
	aggregateSealBytes(digest, []byte(aggregateResultSealDomain))
	aggregateSealInt(digest, int64(limit))
	aggregateSealInt(digest, int64(len(groups)))
	for _, group := range groups {
		aggregateSealBytes(digest, group)
	}
	var resultSeal [sha256.Size]byte
	copy(resultSeal[:], digest.Sum(nil))
	return resultSeal, nil
}

func aggregateSealBytes(digest hash.Hash, value []byte) {
	aggregateSealInt(digest, int64(len(value)))
	_, _ = digest.Write(value)
}

func aggregateSealInt(digest hash.Hash, value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	_, _ = digest.Write(encoded[:])
}
