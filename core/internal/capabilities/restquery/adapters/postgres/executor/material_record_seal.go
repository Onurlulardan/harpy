package executor

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
)

const materialRecordSealDomain = "harpy.restquery.include-material-record.v1"

func sealMaterialRecord(record materialRecord) [sha256.Size]byte {
	digest := sha256.New()
	sealMaterialBytes(digest, []byte(materialRecordSealDomain))
	sealMaterialBytes(digest, record.id[:])
	sealMaterialBytes(digest, record.document)
	sealMaterialUint64(digest, uint64(len(record.order)))
	for _, value := range record.order {
		sealMaterialBytes(digest, value)
	}
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func sealMaterialBytes(digest hash.Hash, value []byte) {
	sealMaterialUint64(digest, uint64(len(value)))
	_, _ = digest.Write(value)
}

func sealMaterialUint64(digest hash.Hash, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = digest.Write(encoded[:])
}
