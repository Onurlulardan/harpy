package profilebackfill

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"strconv"

	"github.com/google/uuid"
)

type recoveryPageEvidence struct {
	afterResourceID      uuid.UUID
	throughResourceID    uuid.UUID
	expectedCount        int
	sourceSHA256         [sha256.Size]byte
	expectedTargetSHA256 [sha256.Size]byte
}

func newRecoveryPageEvidence(
	after uuid.UUID,
	resources []RecoveryResource,
) (recoveryPageEvidence, error) {
	if len(resources) == 0 || len(resources) > MaximumPageSize {
		return recoveryPageEvidence{}, ErrInvalid
	}
	sourceDigest, targetDigest := sha256.New(), sha256.New()
	for _, resource := range resources {
		if !validRecoveryResource(resource) {
			return recoveryPageEvidence{}, ErrInvalid
		}
		writeRecoveryDigest(sourceDigest, []byte(resource.ID.String()),
			[]byte(resource.SchemaVersionID.String()),
			[]byte(strconv.FormatInt(resource.Revision, 10)), resource.SourceSHA256[:])
		writeRecoveryDigest(targetDigest, []byte(resource.ID.String()),
			resource.ExpectedTargetSHA256[:])
	}
	result := recoveryPageEvidence{
		afterResourceID:   after,
		throughResourceID: resources[len(resources)-1].ID,
		expectedCount:     len(resources),
	}
	copy(result.sourceSHA256[:], sourceDigest.Sum(nil))
	copy(result.expectedTargetSHA256[:], targetDigest.Sum(nil))
	return result, nil
}

func rollRecoveryTargetDigest(
	rolling [sha256.Size]byte,
	resource SourceResource,
) [sha256.Size]byte {
	entry := sha256.New()
	writeRecoveryDigest(entry, []byte(resource.ID.String()),
		[]byte(resource.SchemaVersionID.String()),
		[]byte(strconv.FormatInt(resource.Revision, 10)), resource.SHA256[:])
	entrySHA := entry.Sum(nil)
	payload := make([]byte, 0, sha256.Size*2)
	payload = append(payload, rolling[:]...)
	payload = append(payload, entrySHA...)
	clear(entrySHA)
	result := sha256.Sum256(payload)
	clear(payload)
	return result
}

func writeRecoveryDigest(digest hash.Hash, values ...[]byte) {
	var size [8]byte
	for _, value := range values {
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = digest.Write(size[:])
		_, _ = digest.Write(value)
	}
}
