package backupcrypto

import "encoding/binary"

func chunkNonce(base []byte, sequence uint64) ([]byte, error) {
	if len(base) != 12 || sequence == 0 {
		return nil, ErrInvalidArchive
	}
	result := append([]byte(nil), base...)
	current := binary.BigEndian.Uint64(result[4:])
	binary.BigEndian.PutUint64(result[4:], current^sequence)
	return result, nil
}

func chunkAAD(headerDigest []byte, sequence uint64, plaintextBytes uint32) []byte {
	result := make([]byte, len(headerDigest)+12)
	copy(result, headerDigest)
	binary.BigEndian.PutUint64(result[len(headerDigest):], sequence)
	binary.BigEndian.PutUint32(result[len(headerDigest)+8:], plaintextBytes)
	return result
}
