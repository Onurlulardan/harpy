package backupcrypto

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"
)

func Decrypt(
	ciphertext io.Reader,
	plaintext io.Writer,
	keys OpenKeys,
) (int64, error) {
	return decrypt(ciphertext, plaintext, Context{}, keys, false)
}

func DecryptExpected(
	ciphertext io.Reader,
	plaintext io.Writer,
	expected Context,
	keys OpenKeys,
) (int64, error) {
	if !validContext(expected) {
		return 0, ErrInvalidArchive
	}
	return decrypt(ciphertext, plaintext, expected, keys, true)
}

func decrypt(
	ciphertext io.Reader,
	plaintext io.Writer,
	expected Context,
	keys OpenKeys,
	checkContext bool,
) (int64, error) {
	if ciphertext == nil || plaintext == nil {
		return 0, ErrInvalidArchive
	}
	reader := bufio.NewReader(ciphertext)
	headerBytes, header, err := readPreamble(reader)
	if err != nil {
		return 0, err
	}
	if checkContext && header.context() != expected {
		return 0, ErrInvalidArchive
	}
	return decryptBody(reader, plaintext, headerBytes, header, keys)
}

func decryptBody(
	reader *bufio.Reader,
	plaintext io.Writer,
	headerBytes []byte,
	header archiveHeader,
	keys OpenKeys,
) (int64, error) {
	dek, err := unwrapDEK(header, keys)
	if err != nil {
		return 0, err
	}
	defer wipe(dek)
	baseNonce, _ := base64.RawURLEncoding.DecodeString(header.BaseNonce)
	block, _ := aes.NewCipher(dek)
	aead, _ := cipher.NewGCM(block)
	headerDigest := sha256.Sum256(headerBytes)
	total, err := decryptChunks(reader, plaintext, aead, baseNonce, headerDigest[:])
	if err != nil {
		return 0, err
	}
	if _, err = reader.Peek(1); err != io.EOF {
		return 0, ErrInvalidArchive
	}
	return total, nil
}

func readPreamble(reader io.Reader) ([]byte, archiveHeader, error) {
	magic := make([]byte, len(Magic))
	if _, err := io.ReadFull(reader, magic); err != nil || !bytes.Equal(magic, []byte(Magic)) {
		return nil, archiveHeader{}, ErrInvalidArchive
	}
	length, err := readUint32(reader)
	if err != nil || length == 0 || length > MaximumHeaderSize {
		return nil, archiveHeader{}, ErrInvalidArchive
	}
	document := make([]byte, length)
	if _, err = io.ReadFull(reader, document); err != nil {
		return nil, archiveHeader{}, ErrInvalidArchive
	}
	header, err := parseHeader(document)
	return document, header, err
}

func decryptChunks(
	input io.Reader,
	output io.Writer,
	aead cipher.AEAD,
	baseNonce, headerDigest []byte,
) (int64, error) {
	var total int64
	for sequence := uint64(1); ; sequence++ {
		length, err := readUint32(input)
		if err != nil {
			return 0, err
		}
		if length == 0 {
			return total, nil
		}
		if length > ChunkSize {
			return 0, ErrInvalidArchive
		}
		sealed := make([]byte, int(length)+aead.Overhead())
		if _, err = io.ReadFull(input, sealed); err != nil {
			return 0, ErrInvalidArchive
		}
		nonce, _ := chunkNonce(baseNonce, sequence)
		opened, err := aead.Open(nil, nonce, sealed,
			chunkAAD(headerDigest, sequence, length))
		if err != nil {
			return 0, ErrInvalidArchive
		}
		if _, err = output.Write(opened); err != nil {
			return 0, ErrInvalidArchive
		}
		total += int64(length)
	}
}

func readUint32(input io.Reader) (uint32, error) {
	var encoded [4]byte
	if _, err := io.ReadFull(input, encoded[:]); err != nil {
		return 0, ErrInvalidArchive
	}
	return binary.BigEndian.Uint32(encoded[:]), nil
}

func wipe(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
