package fcartv1

import (
	"crypto/cipher"
	"crypto/sha256"
)

func encryptOneChunk(
	sink *ciphertextSink,
	aead cipher.AEAD,
	random RandomSource,
	seen map[[NonceBytes]byte]struct{},
	header Header,
	sequence uint64,
	plaintext []byte,
) error {
	frameBytes := uint64(len(plaintext) + ChunkFramingBytes)
	if frameBytes > sink.limit-sink.written {
		return ErrCiphertextLimit
	}
	nonce, err := random.Nonce()
	if err != nil {
		return ErrRandomSource
	}
	if _, found := seen[nonce]; found {
		return ErrNonceReuse
	}
	seen[nonce] = struct{}{}
	frame, err := sealChunk(aead, header, sequence, plaintext, nonce)
	if err != nil {
		return err
	}
	return sink.writeExact(frame)
}

func encryptResult(
	sink *ciphertextSink,
	header Header,
	plaintextBytes uint64,
	chunkCount uint64,
) EncryptResult {
	encoded, _ := header.encode()
	return EncryptResult{
		CiphertextMetadata: CiphertextMetadata{
			Header: header, HeaderSHA256: sha256.Sum256(encoded[:]),
			CiphertextBytes:  sink.written,
			CiphertextSHA256: sink.digest(),
		},
		PlaintextBytes: plaintextBytes,
		ChunkCount:     chunkCount,
	}
}
