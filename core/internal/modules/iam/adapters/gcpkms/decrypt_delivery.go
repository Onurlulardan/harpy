package gcpkms

import (
	"context"
	"sync"

	"cloud.google.com/go/kms/apiv1/kmspb"
	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/google/uuid"
)

type DeliveryPlaintext struct {
	mutex  sync.Mutex
	value  [deliveryartifactv1.PlaintextBytes]byte
	active bool
}

func (opener *DeliveryOpener) OpenInvitationCredential(
	ctx context.Context,
	tenantID uuid.UUID,
	envelope deliveryartifactv1.Envelope,
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
		TagLength:                         deliveryartifactv1.TagBytes,
		CiphertextCrc32C:                  checksum(sealed),
		AdditionalAuthenticatedDataCrc32C: checksum(envelope.AADSHA256[:]),
		InitializationVectorCrc32C:        checksum(envelope.Nonce[:]),
	})
	if err != nil {
		return nil, ErrUnavailable
	}
	return verifiedDeliveryPlaintext(response)
}

func verifiedDeliveryPlaintext(
	response *kmspb.RawDecryptResponse,
) (*DeliveryPlaintext, error) {
	if response == nil {
		return nil, ErrIntegrity
	}
	defer clear(response.Plaintext)
	if !response.VerifiedCiphertextCrc32C ||
		!response.VerifiedAdditionalAuthenticatedDataCrc32C ||
		!response.VerifiedInitializationVectorCrc32C ||
		len(response.Plaintext) != deliveryartifactv1.PlaintextBytes ||
		!checksumMatches(response.Plaintext, response.PlaintextCrc32C) {
		return nil, ErrIntegrity
	}
	plaintext := &DeliveryPlaintext{active: true}
	copy(plaintext.value[:], response.Plaintext)
	return plaintext, nil
}

func (plaintext *DeliveryPlaintext) RevealOnce(consume func([]byte) error) error {
	if plaintext == nil || consume == nil {
		return ErrContract
	}
	plaintext.mutex.Lock()
	if !plaintext.active {
		plaintext.mutex.Unlock()
		return ErrContract
	}
	var value [deliveryartifactv1.PlaintextBytes]byte
	copy(value[:], plaintext.value[:])
	clear(plaintext.value[:])
	plaintext.active = false
	plaintext.mutex.Unlock()
	err := consume(value[:])
	clear(value[:])
	return err
}

func (plaintext *DeliveryPlaintext) Destroy() {
	if plaintext == nil {
		return
	}
	plaintext.mutex.Lock()
	defer plaintext.mutex.Unlock()
	clear(plaintext.value[:])
	plaintext.active = false
}
