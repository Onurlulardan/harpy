package providerapp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	domainprovider "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/provider"
	"github.com/google/uuid"
)

func webhookAssociatedData(
	receiptID uuid.UUID,
	key domainprovider.Key,
	verified VerifiedWebhook,
) []byte {
	authenticationSHA := verified.AuthenticationSHA256()
	values := []string{
		"billing-webhook-envelope.v1", receiptID.String(), string(key),
		verified.EventID(), verified.AccountID(), verified.ObjectKind(), verified.ObjectID(),
		hex.EncodeToString(authenticationSHA[:]),
	}
	var buffer bytes.Buffer
	for _, value := range values {
		_ = binary.Write(&buffer, binary.BigEndian, uint32(len(value)))
		_, _ = buffer.WriteString(value)
	}
	return buffer.Bytes()
}
