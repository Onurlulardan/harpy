package fcartv1

import "encoding/binary"

const chunkAADBytes = len(Version) + 16 + 8 + 1 + 8 + 4

func encodeChunkAAD(
	header Header,
	sequence uint64,
	plaintextLength uint32,
) ([chunkAADBytes]byte, error) {
	var aad [chunkAADBytes]byte
	if err := header.Validate(); err != nil || plaintextLength == 0 ||
		plaintextLength > header.ChunkSize {
		return aad, ErrChunkLength
	}
	formatCode, err := header.Format.wireCode()
	if err != nil {
		return aad, ErrInvalidHeader
	}
	offset := copy(aad[:], Version)
	offset += copy(aad[offset:], header.ArtifactID[:])
	binary.BigEndian.PutUint64(aad[offset:], header.AttemptGeneration)
	offset += 8
	aad[offset] = formatCode
	offset++
	binary.BigEndian.PutUint64(aad[offset:], sequence)
	offset += 8
	binary.BigEndian.PutUint32(aad[offset:], plaintextLength)
	return aad, nil
}
