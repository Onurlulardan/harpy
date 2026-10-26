package passwordblocklistv1

import (
	"bytes"
	"encoding/binary"
	"sort"
)

const (
	normalizationNFCV1 = 1
	digestSHA256       = 1
)

type Artifact struct {
	kind    Kind
	records []byte
	valid   bool
}

func Parse(content []byte, expected Kind) (Artifact, error) {
	if !expected.Valid() || len(content) < headerSize ||
		!bytes.Equal(content[:8], artifactMagic[:]) ||
		Kind(content[8]) != expected || content[9] != normalizationNFCV1 ||
		content[10] != digestSHA256 || content[11] != 0 {
		return Artifact{}, ErrInvalidArtifact
	}
	count := int(binary.BigEndian.Uint32(content[12:16]))
	if count < 1 || count > expected.MaximumEntries() ||
		len(content) != headerSize+count*digestSize {
		return Artifact{}, ErrInvalidArtifact
	}
	records := content[headerSize:]
	for index := 0; index < count; index++ {
		current := records[index*digestSize : (index+1)*digestSize]
		if allZero(current) || index > 0 &&
			bytes.Compare(records[(index-1)*digestSize:index*digestSize], current) >= 0 {
			return Artifact{}, ErrInvalidArtifact
		}
	}
	return Artifact{kind: expected, records: records, valid: true}, nil
}

func (artifact Artifact) Contains(digest [digestSize]byte) (bool, error) {
	if !artifact.valid || !artifact.kind.Valid() || len(artifact.records)%digestSize != 0 {
		return false, ErrInvalidArtifact
	}
	count := len(artifact.records) / digestSize
	index := sort.Search(count, func(index int) bool {
		record := artifact.records[index*digestSize : (index+1)*digestSize]
		return bytes.Compare(record, digest[:]) >= 0
	})
	return index < count && bytes.Equal(
		artifact.records[index*digestSize:(index+1)*digestSize], digest[:],
	), nil
}

func (artifact Artifact) EntryCount() (int, error) {
	if !artifact.valid || len(artifact.records)%digestSize != 0 {
		return 0, ErrInvalidArtifact
	}
	return len(artifact.records) / digestSize, nil
}

func allZero(value []byte) bool {
	var combined byte
	for _, current := range value {
		combined |= current
	}
	return combined == 0
}
