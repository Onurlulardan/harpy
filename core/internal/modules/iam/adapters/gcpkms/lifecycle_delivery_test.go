package gcpkms

import (
	"bytes"
	"context"
	"errors"
	"testing"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/google/uuid"
)

func TestLifecycleDeliverySealBindsPurposeAndResource(t *testing.T) {
	client := &encryptClientFixture{}
	sealer, err := openDeliverySealer(
		context.Background(), deliveryFixtureKey,
		func(context.Context) (rawEncryptClient, error) { return client, nil },
		bytes.NewReader(bytes.Repeat([]byte{1}, lifecycledeliveryv1.NonceBytes)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer sealer.Close()
	tenantID, artifactID, subjectID := uuid.New(), uuid.New(), uuid.New()
	resourceID, tokenID := uuid.New(), uuid.New()
	envelope, err := sealer.SealLifecycleCredential(
		context.Background(), tenantID, artifactID, subjectID, resourceID, tokenID,
		lifecycledeliveryv1.PurposePasswordReset,
		bytes.Repeat([]byte{7}, lifecycledeliveryv1.PlaintextBytes),
	)
	if err != nil || envelope.Validate(tenantID) != nil {
		t.Fatalf("seal lifecycle delivery: %v", err)
	}
	if client.request == nil || client.request.Name == deliveryFixtureKey {
		t.Fatal("lifecycle delivery did not use the exact KMS resource")
	}
	tampered := envelope
	tampered.Purpose = lifecycledeliveryv1.PurposePortalSetup
	if tampered.Validate(tenantID) == nil {
		t.Fatal("cross-purpose lifecycle delivery was accepted")
	}
}

func TestLifecycleDeliveryOpenRejectsTenantSubstitution(t *testing.T) {
	tenantID, artifactID, subjectID := uuid.New(), uuid.New(), uuid.New()
	resourceID, tokenID := uuid.New(), uuid.New()
	aad, err := lifecycledeliveryv1.AADSHA256(
		tenantID, artifactID, subjectID, resourceID, tokenID,
		lifecycledeliveryv1.PurposeEmailVerification, deliveryFixtureKey,
	)
	if err != nil {
		t.Fatal(err)
	}
	envelope := lifecycledeliveryv1.NewEnvelope(
		artifactID, subjectID, resourceID, tokenID,
		lifecycledeliveryv1.PurposeEmailVerification, deliveryFixtureKey,
		[12]byte{1}, [66]byte{2}, [16]byte{3}, aad,
	)
	client := &decryptClientFixture{response: validDeliveryResponse(bytes.Repeat([]byte{9}, 66))}
	opener, err := openDeliveryOpener(
		context.Background(), func(context.Context) (rawDecryptClient, error) {
			return client, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	defer opener.Close()
	if _, err := opener.OpenLifecycleCredential(
		context.Background(), uuid.New(), envelope,
	); !errors.Is(err, ErrContract) {
		t.Fatalf("cross-tenant lifecycle delivery result = %v", err)
	}
}
