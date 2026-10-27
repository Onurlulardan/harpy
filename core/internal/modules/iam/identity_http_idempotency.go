package iam

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"sort"

	"github.com/google/uuid"
)

const identityOperationNamespace = "harpy/iam/system-resource-operation/v1"

func identityOperationID(
	claims identityEvidenceSource,
	authorityOperation string,
	idempotencyKey string,
) uuid.UUID {
	tenantID, actorID := claims.Tenant(), claims.Subject()
	return identityDeterministicUUID(
		[]byte(identityOperationNamespace), tenantID[:], actorID[:],
		[]byte(authorityOperation), []byte(idempotencyKey),
	)
}

func identityChildID(parent uuid.UUID, purpose string) uuid.UUID {
	return identityDeterministicUUID(
		[]byte(identityOperationNamespace), parent[:], []byte(purpose),
	)
}

func identityDeterministicUUID(parts ...[]byte) uuid.UUID {
	digest := sha256.New()
	var length [4]byte
	for _, part := range parts {
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		_, _ = digest.Write(length[:])
		_, _ = digest.Write(part)
	}
	value := digest.Sum(nil)
	var result uuid.UUID
	copy(result[:], value[:16])
	clear(value)
	result[6] = result[6]&0x0f | 0x80
	result[8] = result[8]&0x3f | 0x80
	return result
}

func identityRequestSHA(value any) ([32]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, ErrIdentityHTTPInvalid
	}
	digest := sha256.Sum256(document)
	clear(document)
	return digest, nil
}

func identitySortedIDs(values []uuid.UUID) []uuid.UUID {
	result := append([]uuid.UUID(nil), values...)
	sort.Slice(result, func(left, right int) bool {
		return result[left].String() < result[right].String()
	})
	return result
}
