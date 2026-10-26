package gcpkms

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/kms/apiv1/kmspb"
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/google/uuid"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/proto"
)

const totpFixtureKey = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
	"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/9"

type totpEncryptFixture struct {
	request  *kmspb.RawEncryptRequest
	response *kmspb.RawEncryptResponse
	invalid  bool
}

func (client *totpEncryptFixture) RawEncrypt(
	_ context.Context,
	request *kmspb.RawEncryptRequest,
	_ ...gax.CallOption,
) (*kmspb.RawEncryptResponse, error) {
	client.request = proto.Clone(request).(*kmspb.RawEncryptRequest)
	sealed := bytes.Repeat([]byte{7}, totpCiphertextBytes+totpTagBytes)
	client.response = &kmspb.RawEncryptResponse{
		Ciphertext: sealed, InitializationVector: bytes.Clone(request.InitializationVector),
		TagLength: totpTagBytes, Name: request.Name,
		CiphertextCrc32C:                          checksum(sealed),
		InitializationVectorCrc32C:                checksum(request.InitializationVector),
		VerifiedPlaintextCrc32C:                   !client.invalid,
		VerifiedAdditionalAuthenticatedDataCrc32C: true,
		VerifiedInitializationVectorCrc32C:        true,
	}
	return client.response, nil
}

func (client *totpEncryptFixture) Close() error { return nil }

func TestTOTPSealUsesPurposeBoundAADAndRawAESGCM(t *testing.T) {
	client := &totpEncryptFixture{}
	sealer, err := openTOTPSealer(
		context.Background(), totpFixtureKey,
		func(context.Context) (rawEncryptClient, error) { return client, nil },
		bytes.NewReader(bytes.Repeat([]byte{1}, totpNonceBytes)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sealer.Close()
	environmentID, tenantID, userID, authenticatorID :=
		uuid.New(), uuid.New(), uuid.New(), uuid.New()
	envelope, err := sealer.SealTOTP(context.Background(), environmentID, tenantID,
		userID, authenticatorID, [totpPlaintextBytes]byte{1})
	if err != nil || envelope.KMSKeyID != totpFixtureKey {
		t.Fatalf("seal TOTP: envelope=%v err=%v", envelope, err)
	}
	wantAAD, _ := TOTPAdditionalAuthenticatedData(
		environmentID, tenantID, userID, authenticatorID, totpFixtureKey,
	)
	deliveryAAD, _ := lifecycledeliveryv1.AADSHA256(
		tenantID, uuid.New(), userID, authenticatorID, uuid.New(),
		lifecycledeliveryv1.PurposePasswordReset, deliveryFixtureKey,
	)
	if client.request == nil ||
		!bytes.Equal(client.request.AdditionalAuthenticatedData, wantAAD[:]) ||
		envelope.AADSHA256 != wantAAD || envelope.AADSHA256 == deliveryAAD ||
		!checksumMatches(client.request.Plaintext, client.request.PlaintextCrc32C) {
		t.Fatal("TOTP KMS purpose or integrity binding drifted")
	}
	if !bytes.Equal(client.response.Ciphertext, make([]byte, totpCiphertextBytes+totpTagBytes)) ||
		!bytes.Equal(client.response.InitializationVector, make([]byte, totpNonceBytes)) {
		t.Fatal("provider response buffers were not wiped")
	}
}

func TestTOTPSealRejectsUnverifiedProviderResponse(t *testing.T) {
	client := &totpEncryptFixture{invalid: true}
	sealer, err := openTOTPSealer(
		context.Background(), totpFixtureKey,
		func(context.Context) (rawEncryptClient, error) { return client, nil },
		bytes.NewReader(bytes.Repeat([]byte{1}, totpNonceBytes)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sealer.Close()
	_, err = sealer.SealTOTP(context.Background(), uuid.New(), uuid.New(), uuid.New(),
		uuid.New(), [totpPlaintextBytes]byte{1})
	if !errors.Is(err, ErrIntegrity) {
		t.Fatalf("unverified response result=%v", err)
	}
}

func TestTOTPAADRejectsCrossPrincipalReplay(t *testing.T) {
	environmentID, tenantID, userID, authenticatorID :=
		uuid.New(), uuid.New(), uuid.New(), uuid.New()
	base, err := TOTPAdditionalAuthenticatedData(
		environmentID, tenantID, userID, authenticatorID, totpFixtureKey,
	)
	if err != nil {
		t.Fatal(err)
	}
	for name, changed := range map[string][32]byte{
		"environment": mustTOTPAAD(t, uuid.New(), tenantID, userID, authenticatorID),
		"tenant":      mustTOTPAAD(t, environmentID, uuid.New(), userID, authenticatorID),
		"user":        mustTOTPAAD(t, environmentID, tenantID, uuid.New(), authenticatorID),
		"authenticator": mustTOTPAAD(
			t, environmentID, tenantID, userID, uuid.New(),
		),
	} {
		if changed == base {
			t.Fatalf("%s substitution preserved AAD", name)
		}
	}
}

func mustTOTPAAD(
	t *testing.T,
	environmentID, tenantID, userID, authenticatorID uuid.UUID,
) [32]byte {
	t.Helper()
	value, err := TOTPAdditionalAuthenticatedData(
		environmentID, tenantID, userID, authenticatorID, totpFixtureKey,
	)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
