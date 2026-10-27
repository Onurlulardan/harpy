package passwordblocklistv1

import (
	"bytes"
	"encoding/binary"
	"sort"
)

func Encode(kind Kind, input [][digestSize]byte) ([]byte, error) {
	if !kind.Valid() || len(input) < 1 || len(input) > kind.MaximumEntries() {
		return nil, ErrInvalidArtifact
	}
	records := append([][digestSize]byte(nil), input...)
	sort.Slice(records, func(left, right int) bool {
		return bytes.Compare(records[left][:], records[right][:]) < 0
	})
	content := make([]byte, headerSize+len(records)*digestSize)
	copy(content[:8], artifactMagic[:])
	content[8] = byte(kind)
	content[9] = normalizationNFCV1
	content[10] = digestSHA256
	binary.BigEndian.PutUint32(content[12:16], uint32(len(records)))
	for index, record := range records {
		if allZero(record[:]) || index > 0 && records[index-1] == record {
			clear(content)
			return nil, ErrInvalidArtifact
		}
		copy(content[headerSize+index*digestSize:], record[:])
	}
	return content, nil
}
