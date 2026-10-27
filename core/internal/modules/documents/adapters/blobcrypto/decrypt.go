package blobcrypto

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
)

func Decrypt(
	ciphertext io.Reader,
	plaintext io.Writer,
	key []byte,
	expected Metadata,
) (Receipt, error) {
	if ciphertext == nil || plaintext == nil || len(key) != 32 || !expected.valid() {
		return Receipt{}, ErrInvalidFormat
	}
	cipherHash := sha256.New()
	source := io.TeeReader(ciphertext, cipherHash)
	metadata, header, err := readHeader(source)
	if err != nil {
		return Receipt{}, err
	}
	if !sameIdentity(metadata, expected) {
		return Receipt{}, ErrIdentityMismatch
	}
	gcm, err := newGCM(key)
	if err != nil {
		return Receipt{}, err
	}
	plainHash := sha256.New()
	target := io.MultiWriter(plaintext, plainHash)
	remaining := metadata.PlaintextBytes
	for expectedIndex := uint32(0); remaining > 0; expectedIndex++ {
		frame := make([]byte, 8)
		if _, err = io.ReadFull(source, frame); err != nil {
			return Receipt{}, ErrInvalidFormat
		}
		index := binary.BigEndian.Uint32(frame)
		size := binary.BigEndian.Uint32(frame[4:])
		if index != expectedIndex || size == 0 || size > ChunkSize || int64(size) > remaining {
			return Receipt{}, ErrInvalidFormat
		}
		nonce := make([]byte, gcm.NonceSize())
		sealed := make([]byte, int(size)+gcm.Overhead())
		if _, err = io.ReadFull(source, nonce); err != nil {
			return Receipt{}, ErrInvalidFormat
		}
		if _, err = io.ReadFull(source, sealed); err != nil {
			return Receipt{}, ErrInvalidFormat
		}
		opened, openErr := gcm.Open(nil, nonce, sealed, chunkAAD(header, index, size))
		if openErr != nil {
			return Receipt{}, errors.Join(ErrContentTampered, openErr)
		}
		if _, err = target.Write(opened); err != nil {
			return Receipt{}, err
		}
		remaining -= int64(size)
	}
	var extra [1]byte
	if count, readErr := source.Read(extra[:]); count != 0 || readErr != io.EOF {
		return Receipt{}, ErrInvalidFormat
	}
	return receipt(metadata, plainHash.Sum(nil), cipherHash.Sum(nil)), nil
}
