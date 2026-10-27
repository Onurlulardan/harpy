package iam

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

const permissionRequestOperationNamespace = "harpy/iam/permission-request-operation/v1"

func permissionRequestOperationID(
	claims accesstoken.Claims,
	operation string,
	idempotencyKey string,
) uuid.UUID {
	tenant, actor := claims.Tenant(), claims.Subject()
	return permissionRequestDeterministicUUID(
		[]byte(permissionRequestOperationNamespace), tenant[:], actor[:],
		[]byte(operation), []byte(idempotencyKey),
	)
}

func permissionRequestChildID(parent uuid.UUID, purpose string) uuid.UUID {
	return permissionRequestDeterministicUUID(
		[]byte(permissionRequestOperationNamespace), parent[:], []byte(purpose),
	)
}

func permissionRequestDeterministicUUID(parts ...[]byte) uuid.UUID {
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

func permissionRequestPurpose(tenantID, groupID uuid.UUID) [32]byte {
	digest := sha256.New()
	_, _ = digest.Write([]byte("harpy/iam/permission-request-decision-purpose/v1"))
	_, _ = digest.Write(tenantID[:])
	_, _ = digest.Write(groupID[:])
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func permissionRequestPurposeWire(tenantID, groupID uuid.UUID) string {
	purpose := permissionRequestPurpose(tenantID, groupID)
	return base64.RawURLEncoding.EncodeToString(purpose[:])
}

func permissionRequestRequestSHA(value any) ([32]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, ErrPermissionRequestHTTPInvalid
	}
	digest := sha256.Sum256(document)
	clear(document)
	return digest, nil
}
