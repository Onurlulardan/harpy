package fcartv1

import (
	"crypto/cipher"
	"io"
)

func Decrypt(
	destination io.Writer,
	ciphertext io.Reader,
	metadata CiphertextMetadata,
	key DEK,
) (DecryptResult, error) {
	if destination == nil || ciphertext == nil {
		return DecryptResult{}, ErrInvalidInput
	}
	if err := metadata.Validate(); err != nil {
		return DecryptResult{}, err
	}
	aead, err := newAEAD(metadata.Header, key)
	if err != nil {
		return DecryptResult{}, err
	}
	source := newCiphertextSource(ciphertext, metadata.CiphertextBytes)
	var encodedHeader [HeaderBytes]byte
	if err := source.readExact(encodedHeader[:]); err != nil {
		return DecryptResult{}, err
	}
	header, err := decodeHeader(encodedHeader)
	if err != nil {
		return DecryptResult{}, err
	}
	if header != metadata.Header {
		return DecryptResult{}, ErrIdentityMismatch
	}
	result, err := decryptChunks(destination, source, aead, header)
	if err != nil {
		return DecryptResult{}, err
	}
	if err := source.requireEOF(); err != nil {
		return DecryptResult{}, err
	}
	if source.digest() != metadata.CiphertextSHA256 {
		return DecryptResult{}, ErrChecksum
	}
	return result, nil
}

func decryptChunks(
	destination io.Writer,
	source *ciphertextSource,
	aead cipher.AEAD,
	header Header,
) (DecryptResult, error) {
	seenNonces := make(map[[NonceBytes]byte]struct{})
	var result DecryptResult
	for source.remaining > 0 {
		length, err := decryptOneChunk(
			destination, source, aead, header, result.ChunkCount, seenNonces,
		)
		if err != nil {
			return DecryptResult{}, err
		}
		result.PlaintextBytes += uint64(length)
		result.ChunkCount++
	}
	return result, nil
}
