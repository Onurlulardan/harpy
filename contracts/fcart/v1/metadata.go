package fcartv1

import "crypto/sha256"

type CiphertextMetadata struct {
	Header           Header
	HeaderSHA256     [32]byte
	CiphertextBytes  uint64
	CiphertextSHA256 [32]byte
}

type EncryptResult struct {
	CiphertextMetadata
	PlaintextBytes uint64
	ChunkCount     uint64
}

type DecryptResult struct {
	PlaintextBytes uint64
	ChunkCount     uint64
}

func (metadata CiphertextMetadata) Validate() error {
	encoded, err := metadata.Header.encode()
	if err != nil || metadata.HeaderSHA256 != sha256.Sum256(encoded[:]) ||
		metadata.CiphertextBytes < uint64(HeaderBytes) ||
		metadata.CiphertextBytes > MaxCiphertextBytes ||
		metadata.CiphertextSHA256 == [32]byte{} {
		return ErrInvalidMetadata
	}
	return nil
}
