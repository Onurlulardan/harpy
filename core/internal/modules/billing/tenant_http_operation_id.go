package billing

import (
	"crypto/sha256"
	"encoding/binary"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

const operationIdentityNamespace = "harpy/billing/tenant-http-operation/v1"

func operationIdentity(
	access billingapi.Access,
	operation string,
	idempotencyKey string,
) (uuid.UUID, error) {
	if !access.ValidFor(billingapi.PermissionBillingManage) ||
		!validIdempotencyKey(idempotencyKey) || operation == "" {
		return uuid.Nil, billingapi.ErrInvalid
	}
	tenantID, principalID := access.TenantID(), access.PrincipalID()
	digest := sha256.New()
	var length [4]byte
	for _, part := range [][]byte{
		[]byte(operationIdentityNamespace), tenantID[:], principalID[:],
		[]byte(operation), []byte(idempotencyKey),
	} {
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		_, _ = digest.Write(length[:])
		_, _ = digest.Write(part)
	}
	material := digest.Sum(nil)
	var result uuid.UUID
	copy(result[:], material[:16])
	clear(material)
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result, nil
}

func validIdempotencyKey(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if character < 0x21 || character > 0x7e {
			return false
		}
	}
	return true
}
