package gcpkms

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/proto"
)

const deliveryFixtureKey = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
	"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/4"

type encryptClientFixture struct {
	request  *kmspb.RawEncryptRequest
	response *kmspb.RawEncryptResponse
	invalid  bool
	closed   int
}

func (client *encryptClientFixture) RawEncrypt(
	_ context.Context,
	request *kmspb.RawEncryptRequest,
	_ ...gax.CallOption,
) (*kmspb.RawEncryptResponse, error) {
	client.request = proto.Clone(request).(*kmspb.RawEncryptRequest)
	sealed := append(
		bytes.Repeat([]byte{2}, deliveryartifactv1.CiphertextBytes),
		bytes.Repeat([]byte{3}, deliveryartifactv1.TagBytes)...,
	)
	client.response = &kmspb.RawEncryptResponse{
		Ciphertext: sealed, InitializationVector: bytes.Clone(request.InitializationVector),
		TagLength: deliveryartifactv1.TagBytes, Name: request.Name,
		CiphertextCrc32C:                          checksum(sealed),
		InitializationVectorCrc32C:                checksum(request.InitializationVector),
		VerifiedPlaintextCrc32C:                   !client.invalid,
		VerifiedAdditionalAuthenticatedDataCrc32C: true,
		VerifiedInitializationVectorCrc32C:        true,
	}
	return client.response, nil
}

func (client *encryptClientFixture) Close() error { client.closed++; return nil }

func TestDeliverySealUsesExactRawAESGCMContractAndWipesProviderBuffers(t *testing.T) {
	client := &encryptClientFixture{}
	sealer, err := openDeliverySealer(
		context.Background(), deliveryFixtureKey,
		func(context.Context) (rawEncryptClient, error) { return client, nil },
		bytes.NewReader(bytes.Repeat([]byte{1}, deliveryartifactv1.NonceBytes)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sealer.Close() })
	tenantID, artifactID, invitationID, tokenID :=
		uuid.New(), uuid.New(), uuid.New(), uuid.New()
	plaintext := bytes.Repeat([]byte{7}, deliveryartifactv1.PlaintextBytes)
	envelope, err := sealer.SealInvitationCredential(
		context.Background(), tenantID, artifactID, invitationID, tokenID, plaintext,
	)
	if err != nil || envelope.Validate(tenantID) != nil {
		t.Fatalf("seal delivery: %v", err)
	}
	request := client.request
	if request == nil || request.Name == deliveryFixtureKey ||
		!checksumMatches(request.Plaintext, request.PlaintextCrc32C) ||
		!checksumMatches(request.AdditionalAuthenticatedData,
			request.AdditionalAuthenticatedDataCrc32C) ||
		!checksumMatches(request.InitializationVector,
			request.InitializationVectorCrc32C) {
		t.Fatal("raw encrypt request contract drifted")
	}
	if !bytes.Equal(client.response.Ciphertext, make([]byte, 82)) ||
		!bytes.Equal(client.response.InitializationVector, make([]byte, 12)) {
		t.Fatal("provider response buffers were not wiped")
	}
}

func TestDeliverySealRejectsUnverifiedProviderResponse(t *testing.T) {
	client := &encryptClientFixture{invalid: true}
	sealer, err := openDeliverySealer(
		context.Background(), deliveryFixtureKey,
		func(context.Context) (rawEncryptClient, error) { return client, nil },
		bytes.NewReader(bytes.Repeat([]byte{1}, 12)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sealer.Close()
	_, err = sealer.SealInvitationCredential(
		context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New(),
		bytes.Repeat([]byte{7}, 66),
	)
	if !errors.Is(err, ErrIntegrity) {
		t.Fatalf("unverified response result = %v", err)
	}
}

func TestDeliveryOpenUsesEnvelopeKeyAndOneTimeWipe(t *testing.T) {
	tenantID, artifactID, invitationID, tokenID :=
		uuid.New(), uuid.New(), uuid.New(), uuid.New()
	aad, _ := deliveryartifactv1.AADSHA256(
		tenantID, artifactID, invitationID, tokenID, deliveryFixtureKey,
	)
	envelope := deliveryartifactv1.NewEnvelope(
		artifactID, invitationID, tokenID, deliveryFixtureKey,
		[12]byte{1}, [66]byte{2}, [16]byte{3}, aad,
	)
	providerPlaintext := bytes.Repeat([]byte{9}, 66)
	client := &decryptClientFixture{response: validDeliveryResponse(providerPlaintext)}
	opener, err := openDeliveryOpener(
		context.Background(), func(context.Context) (rawDecryptClient, error) {
			return client, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer opener.Close()
	plaintext, err := opener.OpenInvitationCredential(
		context.Background(), tenantID, envelope,
	)
	if err != nil {
		t.Fatal(err)
	}
	var revealed []byte
	if err := plaintext.RevealOnce(func(value []byte) error {
		revealed = value
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(revealed, make([]byte, 66)) ||
		!bytes.Equal(client.response.Plaintext, make([]byte, 66)) {
		t.Fatal("delivery plaintext was not wiped")
	}
	if _, err := opener.OpenInvitationCredential(
		context.Background(), uuid.New(), envelope,
	); !errors.Is(err, ErrContract) {
		t.Fatalf("cross-tenant envelope result = %v", err)
	}
}

func validDeliveryResponse(plaintext []byte) *kmspb.RawDecryptResponse {
	return &kmspb.RawDecryptResponse{
		Plaintext: bytes.Clone(plaintext), PlaintextCrc32C: checksum(plaintext),
		VerifiedCiphertextCrc32C:                  true,
		VerifiedAdditionalAuthenticatedDataCrc32C: true,
		VerifiedInitializationVectorCrc32C:        true,
	}
}

func TestDeliveryPlaintextAllowsExactlyOneConcurrentReveal(t *testing.T) {
	plaintext := &DeliveryPlaintext{active: true}
	for index := range plaintext.value {
		plaintext.value[index] = 7
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var consumers atomic.Int32
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			results <- plaintext.RevealOnce(func(value []byte) error {
				if bytes.Equal(value, bytes.Repeat([]byte{7}, len(value))) {
					consumers.Add(1)
				}
				return nil
			})
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !errors.Is(err, ErrContract) {
			t.Fatalf("unexpected reveal error: %v", err)
		}
	}
	if successes != 1 || consumers.Load() != 1 {
		t.Fatalf("successes=%d consumers=%d", successes, consumers.Load())
	}
}
