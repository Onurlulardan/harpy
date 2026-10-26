package blobcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/binary"
	"io"
)

func Encrypt(
	plaintext io.Reader,
	ciphertext io.Writer,
	key []byte,
	metadata Metadata,
	randomness io.Reader,
) (Receipt, error) {
	if plaintext == nil || ciphertext == nil || randomness == nil || len(key) != 32 {
		return Receipt{}, ErrInvalidFormat
	}
	header, err := encodeHeader(metadata)
	if err != nil {
		return Receipt{}, err
	}
	gcm, err := newGCM(key)
	if err != nil {
		return Receipt{}, err
	}
	plainHash := sha256.New()
	cipherHash := sha256.New()
	target := io.MultiWriter(ciphertext, cipherHash)
	if _, err = target.Write(header); err != nil {
		return Receipt{}, err
	}
	buffer := make([]byte, ChunkSize)
	remaining := metadata.PlaintextBytes
	for index := uint32(0); remaining > 0; index++ {
		size := int64(len(buffer))
		if remaining < size {
			size = remaining
		}
		chunk := buffer[:int(size)]
		if _, err = io.ReadFull(plaintext, chunk); err != nil {
			return Receipt{}, ErrInvalidFormat
		}
		if _, err = plainHash.Write(chunk); err != nil {
			return Receipt{}, err
		}
		nonce := make([]byte, gcm.NonceSize())
		if _, err = io.ReadFull(randomness, nonce); err != nil {
			return Receipt{}, err
		}
		frame := make([]byte, 8)
		binary.BigEndian.PutUint32(frame, index)
		binary.BigEndian.PutUint32(frame[4:], uint32(size))
		sealed := gcm.Seal(nil, nonce, chunk, chunkAAD(header, index, uint32(size)))
		for _, value := range [][]byte{frame, nonce, sealed} {
			if _, err = target.Write(value); err != nil {
				return Receipt{}, err
			}
		}
		remaining -= size
	}
	var extra [1]byte
	if count, readErr := plaintext.Read(extra[:]); count != 0 || readErr != io.EOF {
		return Receipt{}, ErrInvalidFormat
	}
	return receipt(metadata, plainHash.Sum(nil), cipherHash.Sum(nil)), nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrInvalidFormat
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil || gcm.NonceSize() != 12 {
		return nil, ErrInvalidFormat
	}
	return gcm, nil
}

func receipt(metadata Metadata, plaintext, ciphertext []byte) Receipt {
	result := Receipt{Metadata: metadata}
	copy(result.PlaintextSHA256[:], plaintext)
	copy(result.CiphertextSHA256[:], ciphertext)
	return result
}
