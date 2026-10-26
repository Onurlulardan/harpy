package backupcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"io"
)

func Encrypt(
	plaintext io.Reader,
	ciphertext io.Writer,
	context Context,
	keys SealKeys,
) (Result, error) {
	if plaintext == nil || ciphertext == nil || !validContext(context) || keys.Random == nil {
		return Result{}, ErrInvalidArchive
	}
	dek := make([]byte, 32)
	baseNonce := make([]byte, 12)
	if _, err := io.ReadFull(keys.Random, dek); err != nil {
		return Result{}, ErrInvalidArchive
	}
	defer wipe(dek)
	if _, err := io.ReadFull(keys.Random, baseNonce); err != nil {
		return Result{}, ErrInvalidArchive
	}
	drWrap, recoveryWrap, err := wrapDEK(context, dek, keys)
	if err != nil {
		return Result{}, err
	}
	header := archiveHeader{Schema: HeaderSchema, EnvironmentID: context.EnvironmentID,
		RecoverySetID: context.RecoverySetID, ComponentID: context.ComponentID,
		ChunkSize: ChunkSize, BaseNonce: base64.RawURLEncoding.EncodeToString(baseNonce),
		DRWrap: drWrap, RecoveryWrap: recoveryWrap}
	headerBytes, err := canonicalHeader(header)
	if err != nil {
		return Result{}, err
	}
	hash := sha256.New()
	counting := &countingWriter{writer: io.MultiWriter(ciphertext, hash)}
	if err = writePreamble(counting, headerBytes); err != nil {
		return Result{}, err
	}
	plainBytes, err := encryptChunks(plaintext, counting, dek, baseNonce, headerBytes)
	if err != nil {
		return Result{}, err
	}
	return Result{CiphertextBytes: counting.count,
		CiphertextSHA256: hex.EncodeToString(hash.Sum(nil)), PlaintextBytes: plainBytes}, nil
}

func encryptChunks(input io.Reader, output io.Writer, dek, baseNonce, header []byte) (int64, error) {
	block, err := aes.NewCipher(dek)
	if err != nil {
		return 0, ErrInvalidArchive
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return 0, ErrInvalidArchive
	}
	headerDigest := sha256.Sum256(header)
	buffer := make([]byte, ChunkSize)
	var total int64
	for sequence := uint64(1); ; sequence++ {
		read, readErr := io.ReadFull(input, buffer)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return 0, ErrInvalidArchive
		}
		if read == 0 {
			return total, writeUint32(output, 0)
		}
		nonce, _ := chunkNonce(baseNonce, sequence)
		sealed := aead.Seal(nil, nonce, buffer[:read],
			chunkAAD(headerDigest[:], sequence, uint32(read)))
		if err = writeUint32(output, uint32(read)); err != nil {
			return 0, err
		}
		if _, err = output.Write(sealed); err != nil {
			return 0, ErrInvalidArchive
		}
		total += int64(read)
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			return total, writeUint32(output, 0)
		}
	}
}

func writePreamble(output io.Writer, header []byte) error {
	if _, err := output.Write([]byte(Magic)); err != nil || writeUint32(output, uint32(len(header))) != nil {
		return ErrInvalidArchive
	}
	_, err := output.Write(header)
	if err != nil {
		return ErrInvalidArchive
	}
	return nil
}

func writeUint32(output io.Writer, value uint32) error {
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	_, err := output.Write(encoded[:])
	if err != nil {
		return ErrInvalidArchive
	}
	return nil
}

type countingWriter struct {
	writer io.Writer
	count  int64
}

func (writer *countingWriter) Write(value []byte) (int, error) {
	written, err := writer.writer.Write(value)
	writer.count += int64(written)
	return written, err
}
