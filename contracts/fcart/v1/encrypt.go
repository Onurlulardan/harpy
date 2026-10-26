package fcartv1

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
)

func Encrypt(
	destination io.Writer,
	plaintext io.Reader,
	header Header,
	key DEK,
	random RandomSource,
) (EncryptResult, error) {
	return encryptWithLimit(
		destination, plaintext, header, key, random, MaxCiphertextBytes,
	)
}

func encryptWithLimit(
	destination io.Writer,
	plaintext io.Reader,
	header Header,
	key DEK,
	random RandomSource,
	limit uint64,
) (EncryptResult, error) {
	if destination == nil || plaintext == nil {
		return EncryptResult{}, ErrInvalidInput
	}
	if random == nil {
		return EncryptResult{}, ErrRandomSource
	}
	aead, err := newAEAD(header, key)
	if err != nil {
		return EncryptResult{}, err
	}
	sink := newCiphertextSink(destination, limit)
	encodedHeader, err := header.encode()
	if err != nil {
		return EncryptResult{}, err
	}
	if err := sink.writeExact(encodedHeader[:]); err != nil {
		return EncryptResult{}, err
	}
	return encryptChunks(sink, plaintext, header, aead, random)
}

func encryptChunks(
	sink *ciphertextSink,
	plaintext io.Reader,
	header Header,
	aead cipher.AEAD,
	random RandomSource,
) (EncryptResult, error) {
	buffer := make([]byte, ChunkPlaintextBytes)
	defer clear(buffer)
	seenNonces := make(map[[NonceBytes]byte]struct{})
	var plaintextBytes, sequence uint64
	for {
		length, readErr := io.ReadFull(plaintext, buffer)
		if errors.Is(readErr, io.EOF) {
			return encryptResult(sink, header, plaintextBytes, sequence), nil
		}
		if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
			return EncryptResult{}, errors.Join(ErrPlaintextRead, readErr)
		}
		if length == 0 {
			return encryptResult(sink, header, plaintextBytes, sequence), nil
		}
		if err := encryptOneChunk(
			sink, aead, random, seenNonces, header, sequence, buffer[:length],
		); err != nil {
			return EncryptResult{}, err
		}
		clear(buffer[:length])
		plaintextBytes += uint64(length)
		sequence++
		if errors.Is(readErr, io.ErrUnexpectedEOF) {
			return encryptResult(sink, header, plaintextBytes, sequence), nil
		}
	}
}

func newAEAD(header Header, key DEK) (cipher.AEAD, error) {
	if err := header.Validate(); err != nil {
		return nil, err
	}
	if err := key.Validate(); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, ErrInvalidDEK
	}
	return cipher.NewGCM(block)
}
