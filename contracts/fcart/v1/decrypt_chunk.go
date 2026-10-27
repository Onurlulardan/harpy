package fcartv1

import (
	"crypto/cipher"
	"encoding/binary"
	"io"
)

func decryptOneChunk(
	destination io.Writer,
	source *ciphertextSource,
	aead cipher.AEAD,
	header Header,
	expectedSequence uint64,
	seenNonces map[[NonceBytes]byte]struct{},
) (uint32, error) {
	if source.remaining < ChunkFramingBytes+1 {
		return 0, ErrTruncated
	}
	var prefix [ChunkFramePrefixBytes]byte
	if err := source.readExact(prefix[:]); err != nil {
		return 0, err
	}
	sequence := binary.BigEndian.Uint64(prefix[:8])
	if sequence != expectedSequence {
		return 0, ErrSequence
	}
	length := binary.BigEndian.Uint32(prefix[8:12])
	if length == 0 || length > header.ChunkSize {
		return 0, ErrChunkLength
	}
	sealedBytes := uint64(length) + TagBytes
	if source.remaining < sealedBytes {
		return 0, ErrTruncated
	}
	if source.remaining > sealedBytes && length != header.ChunkSize {
		return 0, ErrChunkLength
	}
	var nonce [NonceBytes]byte
	copy(nonce[:], prefix[12:])
	if _, duplicate := seenNonces[nonce]; duplicate {
		return 0, ErrNonceReuse
	}
	seenNonces[nonce] = struct{}{}
	sealed := make([]byte, sealedBytes)
	if err := source.readExact(sealed); err != nil {
		return 0, err
	}
	aad, err := encodeChunkAAD(header, sequence, length)
	if err != nil {
		return 0, err
	}
	plaintext, err := aead.Open(nil, nonce[:], sealed, aad[:])
	if err != nil {
		return 0, ErrAuthentication
	}
	defer clear(plaintext)
	if err := writePlaintextExact(destination, plaintext); err != nil {
		return 0, err
	}
	return length, nil
}
