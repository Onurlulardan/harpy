package billing

import (
	"crypto/sha256"
	"encoding/binary"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
	"github.com/google/uuid"
)

const platformReportOperationNamespace = "harpy/billing/platform-report-operation/v1"

func platformReportOperationID(
	access billingapi.Access,
	subscriptionID uuid.UUID,
	idempotencyKey string,
) (uuid.UUID, error) {
	if !access.ValidFor(billingapi.PermissionPlatformBillingManage) ||
		access.Scope() != billingapi.AccessScopeGlobal || subscriptionID == uuid.Nil ||
		!validIdempotencyKey(idempotencyKey) {
		return uuid.Nil, billingapi.ErrInvalid
	}
	digest := sha256.New()
	var length [4]byte
	principalID := access.PrincipalID()
	for _, part := range [][]byte{
		[]byte(platformReportOperationNamespace), principalID[:],
		subscriptionID[:], []byte(operationSubscriptionTransition),
		[]byte(idempotencyKey),
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
