package fcartv1

import (
	"crypto/cipher"
	"encoding/binary"
)

func sealChunk(
	aead cipher.AEAD,
	header Header,
	sequence uint64,
	plaintext []byte,
	nonce [NonceBytes]byte,
) ([]byte, error) {
	length := uint32(len(plaintext))
	aad, err := encodeChunkAAD(header, sequence, length)
	if err != nil {
		return nil, err
	}
	frame := make(
		[]byte, ChunkFramePrefixBytes,
		ChunkFramePrefixBytes+len(plaintext)+TagBytes,
	)
	binary.BigEndian.PutUint64(frame[:8], sequence)
	binary.BigEndian.PutUint32(frame[8:12], length)
	copy(frame[12:], nonce[:])
	return aead.Seal(frame, nonce[:], plaintext, aad[:]), nil
}
