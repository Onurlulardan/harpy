package recordpage

import (
	"crypto/sha256"
	"encoding/binary"
)

const pageSealDomain = "harpy.restquery.record-page.v1"

func cloneRecords(records [][]byte) [][]byte {
	if records == nil {
		return nil
	}
	result := make([][]byte, len(records))
	for index := range records {
		result[index] = append([]byte(nil), records[index]...)
	}
	return result
}

func sealPage(page Page) [sha256.Size]byte {
	hash := sha256.New()
	writeBytes(hash.Write, []byte(pageSealDomain))
	writeUint64(hash.Write, uint64(page.requestedSize))
	if page.hasMore {
		_, _ = hash.Write([]byte{1})
	} else {
		_, _ = hash.Write([]byte{0})
	}
	writeBytes(hash.Write, []byte(page.nextCursor))
	writeUint64(hash.Write, uint64(len(page.records)))
	for _, record := range page.records {
		writeBytes(hash.Write, record)
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func writeBytes(write func([]byte) (int, error), value []byte) {
	writeUint64(write, uint64(len(value)))
	_, _ = write(value)
}

func writeUint64(write func([]byte) (int, error), value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = write(encoded[:])
}
