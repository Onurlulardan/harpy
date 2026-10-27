package installationdelivery

import (
	"context"
	"strings"
	"testing"

	credentialdeliveryv1 "github.com/Onurlulardan/harpy/contracts/credentialdelivery/v1"
	"github.com/google/uuid"
)

func TestWP10CredentialDeliveryAcknowledgesDeletionExactlyOnce(t *testing.T) {
	ctx := context.Background()
	operation, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	attempt, credential := uuid.New(), uuid.New()
	store := &credentialStore{}
	ready, err := Stage(ctx, store, StageCommand{
		OperationID: operation, AttemptID: attempt, CredentialID: credential,
		CredentialVersion:          1,
		InstallationManifestSHA256: strings.Repeat("a", 64),
		DeliveryManifestSHA256:     credentialdeliveryv1.ManifestSHA256(),
	})
	if err != nil {
		t.Fatal(err)
	}
	token := strings.Repeat("b", 64)
	claimed, err := Claim(ctx, store, ClaimCommand{
		OperationID: operation, AttemptID: attempt,
		ExpectedRevision: ready.Revision, AdapterID: "local-owner-file",
		ClaimTokenSHA256: token, LeaseSeconds: 60,
	})
	if err != nil {
		t.Fatal(err)
	}
	acknowledged, err := Acknowledge(ctx, store, AcknowledgeCommand{
		OperationID: operation, AttemptID: attempt,
		ExpectedRevision: claimed.Revision, AdapterID: claimed.ClaimAdapterID,
		ClaimTokenSHA256: token,
	})
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := DeleteSecret(ctx, store, StateCommand{
		OperationID: operation, AttemptID: attempt,
		ExpectedRevision: acknowledged.Revision,
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, err := Complete(ctx, store, StateCommand{
		OperationID: operation, AttemptID: attempt,
		ExpectedRevision: deleted.Revision,
	})
	if err != nil || completed.State != StateCompleted ||
		store.value.SecretDeletedAt == nil {
		t.Fatalf("complete = %#v delivery=%#v err=%v", completed, store.value, err)
	}
}
