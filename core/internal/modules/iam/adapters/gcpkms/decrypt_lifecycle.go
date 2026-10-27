package gcpkms

import (
	"context"

	"cloud.google.com/go/kms/apiv1/kmspb"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/google/uuid"
)

func (opener *DeliveryOpener) OpenLifecycleCredential(
	ctx context.Context,
	tenantID uuid.UUID,
	envelope lifecycledeliveryv1.Envelope,
) (*DeliveryPlaintext, error) {
	if opener == nil || opener.client == nil || opener.closed.Load() ||
		ctx == nil || ctx.Err() != nil || envelope.Validate(tenantID) != nil {
		return nil, ErrContract
	}
	key, err := parseKeyVersion(envelope.KeyID)
	if err != nil {
		return nil, ErrContract
	}
	resource, err := key.Resource()
	if err != nil {
		return nil, ErrContract
	}
	sealed := make([]byte, 0, len(envelope.Ciphertext)+len(envelope.Tag))
	sealed = append(sealed, envelope.Ciphertext[:]...)
	sealed = append(sealed, envelope.Tag[:]...)
	defer clear(sealed)
	response, err := opener.client.RawDecrypt(ctx, &kmspb.RawDecryptRequest{
		Name: resource, Ciphertext: sealed,
		AdditionalAuthenticatedData:       envelope.AADSHA256[:],
		InitializationVector:              envelope.Nonce[:],
		TagLength:                         lifecycledeliveryv1.TagBytes,
		CiphertextCrc32C:                  checksum(sealed),
		AdditionalAuthenticatedDataCrc32C: checksum(envelope.AADSHA256[:]),
		InitializationVectorCrc32C:        checksum(envelope.Nonce[:]),
	})
	if err != nil {
		return nil, ErrUnavailable
	}
	return verifiedDeliveryPlaintext(response)
}
