package domain

const (
	cipherHeaderBytes = int64(74)
	cipherFrameBytes  = int64(36)
	cipherChunkBytes  = int64(8 * 1024 * 1024)
)

func CiphertextUpperBound(plaintextBytes int64) (int64, error) {
	return ciphertextUpperBound(plaintextBytes, MaxUploadBytes)
}

func ArchiveCiphertextUpperBound(plaintextBytes int64) (int64, error) {
	return ciphertextUpperBound(plaintextBytes, MaximumArchiveArtifactPlaintext)
}

func ciphertextUpperBound(plaintextBytes, maximum int64) (int64, error) {
	if plaintextBytes < 1 || plaintextBytes > maximum {
		return 0, ErrUploadInvalid
	}
	chunks := (plaintextBytes + cipherChunkBytes - 1) / cipherChunkBytes
	return cipherHeaderBytes + plaintextBytes + chunks*cipherFrameBytes, nil
}
