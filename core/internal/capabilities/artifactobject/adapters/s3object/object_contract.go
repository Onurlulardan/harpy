package s3object

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	metadataContract = "harpy-contract"
	metadataDigest   = "harpy-ciphertext-sha256"
	metadataHeader   = "harpy-header-sha256"
	metadataBytes    = "harpy-ciphertext-bytes"
)

func validToken(value string, maximum int, lower bool) bool {
	if len(value) < 1 || len(value) > maximum || strings.TrimSpace(value) != value {
		return false
	}
	for index := range len(value) {
		character := value[index]
		letter := character >= 'a' && character <= 'z'
		if !lower {
			letter = letter || character >= 'A' && character <= 'Z'
		}
		if !letter && (character < '0' || character > '9') &&
			character != '-' && character != '_' && character != '.' {
			return false
		}
	}
	return true
}

func (store *Store) validObjectKey(key string) bool {
	parts := strings.Split(key, "/")
	if len(parts) != 4 || parts[0] != store.prefix || parts[1] != "v1" ||
		!strings.HasSuffix(parts[3], ".fcart") {
		return false
	}
	id, err := uuid.Parse(parts[2])
	if err != nil || id.Version() != 7 || strings.ToLower(id.String()) != parts[2] {
		return false
	}
	generation, err := strconv.Atoi(strings.TrimSuffix(parts[3], ".fcart"))
	return err == nil && generation >= 1 && generation <= 3
}

func (store *Store) objectMetadata(ciphertext, header [sha256.Size]byte,
	size int64) map[string]string {
	return map[string]string{metadataContract: store.contract,
		metadataDigest: hex.EncodeToString(ciphertext[:]),
		metadataHeader: hex.EncodeToString(header[:]),
		metadataBytes:  strconv.FormatInt(size, 10)}
}

func checksumBase64(digest [sha256.Size]byte) string {
	return base64.StdEncoding.EncodeToString(digest[:])
}

func (store *Store) verifyMetadata(metadata map[string]string, checksum *string,
	ciphertext, header [sha256.Size]byte, size int64) error {
	for key, value := range store.objectMetadata(ciphertext, header, size) {
		if metadata[key] != value {
			return fmt.Errorf("%w: metadata %s", ErrProviderEvidence, key)
		}
	}
	if checksum == nil || *checksum != checksumBase64(ciphertext) {
		return fmt.Errorf("%w: SHA-256 checksum", ErrProviderEvidence)
	}
	return nil
}
