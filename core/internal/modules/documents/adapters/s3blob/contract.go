package s3blob

import (
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	metadataContract   = "harpy-contract"
	metadataGeneration = "harpy-object-generation"
	metadataBytes      = "harpy-ciphertext-bytes"
	contractFCDOC1     = "FCDOC1"
)

func validKey(value string) bool {
	parts := strings.Split(value, "/")
	if len(parts) == 5 && parts[0] == "tenants" &&
		parts[2] == "document-archives" && parts[4] == "content" &&
		validTenantToken(parts[1]) {
		id, err := uuid.Parse(parts[3])
		return err == nil && id.Version() == 7 && id.String() == parts[3]
	}
	if len(parts) != 7 || parts[0] != "tenants" || parts[2] != "documents" ||
		parts[4] != "versions" || parts[6] != "content" ||
		!validTenantToken(parts[1]) {
		return false
	}
	for _, part := range []string{parts[3], parts[5]} {
		id, err := uuid.Parse(part)
		if err != nil || id.Version() != 7 || id.String() != part {
			return false
		}
	}
	return true
}

func validTenantToken(value string) bool {
	if len(value) != 26 {
		return false
	}
	for _, character := range value {
		if character < 'a' || character > 'z' {
			if character < '2' || character > '7' {
				return false
			}
		}
	}
	return true
}

func metadata(generation uint64, bytes int64) map[string]string {
	return map[string]string{metadataContract: contractFCDOC1,
		metadataGeneration: strconv.FormatUint(generation, 10),
		metadataBytes:      strconv.FormatInt(bytes, 10)}
}

func checksum(digest [32]byte) string {
	return base64.StdEncoding.EncodeToString(digest[:])
}

func matchingMetadata(actual map[string]string, generation uint64, bytes int64) bool {
	expected := metadata(generation, bytes)
	for key, value := range expected {
		if actual[key] != value {
			return false
		}
	}
	return true
}
